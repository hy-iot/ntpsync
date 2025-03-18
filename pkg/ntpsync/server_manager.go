package ntpsync

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ServerManager 处理多个NTP服务器并提供故障转移功能
type ServerManager struct {
	// servers 是服务器地址到其状态的映射
	servers map[string]*ServerStatus
	
	// serverOrder 是按优先级排序的服务器列表
	serverOrder []string
	
	// mutex 用于线程安全
	mutex sync.RWMutex
	
	// timeout 用于服务器请求的超时时间
	timeout time.Duration
}

// NewServerManager 创建一个新的服务器管理器，使用给定的服务器
func NewServerManager(servers []string, timeout time.Duration) (*ServerManager, error) {
	if len(servers) == 0 {
		return nil, errors.New("必须提供至少一个NTP服务器")
	}
	
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	
	sm := &ServerManager{
		servers:     make(map[string]*ServerStatus),
		serverOrder: make([]string, 0, len(servers)),
		timeout:     timeout,
	}
	
	// 初始化服务器状态
	for _, server := range servers {
		sm.servers[server] = &ServerStatus{
			Address: server,
		}
		sm.serverOrder = append(sm.serverOrder, server)
	}
	
	return sm, nil
}

// AddServer 向管理器添加新服务器
func (sm *ServerManager) AddServer(server string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// 检查服务器是否已存在
	if _, exists := sm.servers[server]; exists {
		return fmt.Errorf("服务器 %s 已存在", server)
	}
	
	// 添加服务器
	sm.servers[server] = &ServerStatus{
		Address: server,
	}
	sm.serverOrder = append(sm.serverOrder, server)
	
	return nil
}

// RemoveServer 从管理器中移除服务器
func (sm *ServerManager) RemoveServer(server string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// 检查服务器是否存在
	if _, exists := sm.servers[server]; !exists {
		return fmt.Errorf("服务器 %s 不存在", server)
	}
	
	// 从映射中移除
	delete(sm.servers, server)
	
	// 从顺序列表中移除
	for i, s := range sm.serverOrder {
		if s == server {
			sm.serverOrder = append(sm.serverOrder[:i], sm.serverOrder[i+1:]...)
			break
		}
	}
	
	return nil
}

// GetServers 返回所有服务器的列表
func (sm *ServerManager) GetServers() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	servers := make([]string, len(sm.serverOrder))
	copy(servers, sm.serverOrder)
	
	return servers
}

// GetServerStatus 返回特定服务器的状态
func (sm *ServerManager) GetServerStatus(server string) (*ServerStatus, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	status, exists := sm.servers[server]
	if !exists {
		return nil, fmt.Errorf("服务器 %s 不存在", server)
	}
	
	// 返回副本以防止外部修改
	statusCopy := *status
	return &statusCopy, nil
}

// GetAllServerStatuses 返回所有服务器的状态
func (sm *ServerManager) GetAllServerStatuses() []ServerStatus {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	statuses := make([]ServerStatus, 0, len(sm.servers))
	
	for _, server := range sm.serverOrder {
		status := *sm.servers[server]
		statuses = append(statuses, status)
	}
	
	return statuses
}

// UpdateServerStatus 更新服务器的状态
func (sm *ServerManager) UpdateServerStatus(server string, status ServerStatus) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	serverStatus, exists := sm.servers[server]
	if !exists {
		return fmt.Errorf("服务器 %s 不存在", server)
	}
	
	*serverStatus = status
	
	// 根据可达性和层级重新排序服务器
	sm.reorderServers()
	
	return nil
}

// reorderServers 根据服务器状态重新排序服务器
// 服务器按以下顺序排序：
// 1. 可达性（可达服务器优先）
// 2. 层级（较低层级优先）
// 3. RTT（较低RTT优先）
func (sm *ServerManager) reorderServers() {
	// 创建服务器地址的切片
	servers := make([]string, 0, len(sm.servers))
	for server := range sm.servers {
		servers = append(servers, server)
	}
	
	// 排序服务器
	sort.SliceStable(servers, func(i, j int) bool {
		si := sm.servers[servers[i]]
		sj := sm.servers[servers[j]]
		
		// 可达服务器优先
		if si.Reachable != sj.Reachable {
			return si.Reachable
		}
		
		// 较低层级优先
		if si.Stratum != sj.Stratum {
			return si.Stratum < sj.Stratum
		}
		
		// 较低RTT优先
		return si.RTT < sj.RTT
	})
	
	// 更新服务器顺序
	sm.serverOrder = servers
}

// GetBestServer 返回最佳服务器
func (sm *ServerManager) GetBestServer() (string, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	if len(sm.serverOrder) == 0 {
		return "", errors.New("没有可用的服务器")
	}
	
	// 查找第一个可达的服务器
	for _, server := range sm.serverOrder {
		if sm.servers[server].Reachable {
			return server, nil
		}
	}
	
	// 如果没有找到可达的服务器，返回第一个
	// 调用者必须处理服务器不可达的情况
	return sm.serverOrder[0], nil
}

// ProbeAllServers 探测所有服务器并更新它们的状态
func (sm *ServerManager) ProbeAllServers(ntpClient *NTPSync) error {
	sm.mutex.RLock()
	servers := make([]string, len(sm.serverOrder))
	copy(servers, sm.serverOrder)
	sm.mutex.RUnlock()
	
	if len(servers) == 0 {
		return errors.New("没有可用的服务器")
	}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var lastErr error
	
	for _, server := range servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			
			result, err := ntpClient.syncWithServerBinary(server, sm.timeout)
			
			status := ServerStatus{
				Address: server,
			}
			
			if err != nil {
				status.Reachable = false
				
				mu.Lock()
				lastErr = err
				mu.Unlock()
			} else {
				status.Reachable = true
				status.LastResponse = time.Now()
				status.RTT = result.RTT
				status.Stratum = result.Stratum
				status.Offset = result.Offset
			}
			
			_ = sm.UpdateServerStatus(server, status)
		}(server)
	}
	
	wg.Wait()
	
	// 检查是否至少有一个服务器可达
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	for _, server := range sm.serverOrder {
		if sm.servers[server].Reachable {
			return nil
		}
	}
	
	// 没有可达的服务器
	if lastErr != nil {
		return fmt.Errorf("所有服务器都不可达: %v", lastErr)
	}
	
	return errors.New("所有服务器都不可达")
}
