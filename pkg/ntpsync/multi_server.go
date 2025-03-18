package ntpsync

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// SyncWithMultiServer 执行与多个NTP服务器的同步
// 按照优先顺序尝试服务器，并使用第一个成功的服务器
func (n *NTPSync) SyncWithMultiServer() error {
	n.mutex.Lock()
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	timeout := n.Timeout
	n.mutex.Unlock()

	if len(servers) == 0 {
		return errors.New("未配置NTP服务器")
	}

	// 按顺序尝试每个服务器
	var lastErr error
	for _, server := range servers {
		result, err := n.syncWithServerBinary(server, timeout)
		if err != nil {
			lastErr = err
			continue
		}

		// 成功与此服务器同步
		n.mutex.Lock()
		n.TimeOffset = result.Offset
		n.LastSync = time.Now()
		n.mutex.Unlock()
		
		return nil
	}

	// 如果执行到这里，说明所有服务器都失败了
	return fmt.Errorf("无法与任何NTP服务器同步: %v", lastErr)
}

// SyncWithMultiServerParallel 并行执行与多个NTP服务器的同步
// 同时尝试所有服务器，并使用响应最快的服务器的结果
func (n *NTPSync) SyncWithMultiServerParallel() error {
	n.mutex.Lock()
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	timeout := n.Timeout
	n.mutex.Unlock()

	if len(servers) == 0 {
		return errors.New("未配置NTP服务器")
	}

	// 创建结果和错误的通道
	resultChan := make(chan *SyncResult, len(servers))
	errChan := make(chan error, len(servers))
	
	// 并行尝试所有服务器
	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			
			result, err := n.syncWithServerBinary(server, timeout)
			if err != nil {
				errChan <- err
				return
			}
			
			resultChan <- result
		}(server)
	}
	
	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()
	
	// 获取第一个成功的结果
	var result *SyncResult
	var lastErr error
	
	// 检查结果
	for r := range resultChan {
		if result == nil || r.Stratum < result.Stratum || (r.Stratum == result.Stratum && r.RTT < result.RTT) {
			result = r
		}
	}
	
	// 如果没有结果，检查错误
	if result == nil {
		for err := range errChan {
			lastErr = err
		}
		
		if lastErr != nil {
			return fmt.Errorf("无法与任何NTP服务器同步: %v", lastErr)
		}
		
		return errors.New("无法与任何NTP服务器同步")
	}
	
	// 成功同步
	n.mutex.Lock()
	n.TimeOffset = result.Offset
	n.LastSync = time.Now()
	n.mutex.Unlock()
	
	return nil
}

// GetMultiServerStatus 返回所有已配置NTP服务器的状态
func (n *NTPSync) GetMultiServerStatus() ([]ServerStatus, error) {
	n.mutex.RLock()
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	timeout := n.Timeout
	n.mutex.RUnlock()

	if len(servers) == 0 {
		return nil, errors.New("未配置NTP服务器")
	}

	// 创建结果通道
	statusChan := make(chan ServerStatus, len(servers))
	
	// 并行检查所有服务器
	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			
			status := ServerStatus{
				Address: server,
			}
			
			result, err := n.syncWithServerBinary(server, timeout)
			if err != nil {
				status.Reachable = false
			} else {
				status.Reachable = true
				status.LastResponse = time.Now()
				status.RTT = result.RTT
				status.Stratum = result.Stratum
				status.Offset = result.Offset
			}
			
			statusChan <- status
		}(server)
	}
	
	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(statusChan)
	}()
	
	// 收集所有状态
	statuses := make([]ServerStatus, 0, len(servers))
	for status := range statusChan {
		statuses = append(statuses, status)
	}
	
	return statuses, nil
}

// UpdateNTPSyncWithMultiServer 更新NTPSync结构体以使用多服务器功能
func (n *NTPSync) UpdateNTPSyncWithMultiServer() {
	// 我们不能直接分配给方法，所以我们将使用一个包装函数
	// 在当前实现中，这是一个空操作
}
