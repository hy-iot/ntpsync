package ntpsync

import (
	"time"
)

// Sync 执行一次与NTP服务器的同步
// 这是对SyncWithBinary的包装，用于向后兼容
func (n *NTPSync) Sync() error {
	return n.SyncWithBinary()
}

// GetStatus 返回所有已配置NTP服务器的状态
// 这是对GetStatusBinary的包装，用于向后兼容
func (n *NTPSync) GetStatus() ([]ServerStatus, error) {
	return n.GetStatusBinary()
}

// Now 返回经NTP偏移量调整后的当前时间
func (n *NTPSync) Now() time.Time {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	return time.Now().Add(n.TimeOffset)
}

// LastSyncTime 返回最后一次成功同步的时间
func (n *NTPSync) LastSyncTime() time.Time {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	return n.LastSync
}

// TimeOffsetDuration 返回当前与NTP服务器的时间偏移量
func (n *NTPSync) TimeOffsetDuration() time.Duration {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	return n.TimeOffset
}

// AddServer 向列表中添加新的NTP服务器
func (n *NTPSync) AddServer(server string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	
	// 检查服务器是否已存在
	for _, s := range n.Servers {
		if s == server {
			return
		}
	}
	
	n.Servers = append(n.Servers, server)
}

// RemoveServer 从列表中移除NTP服务器
func (n *NTPSync) RemoveServer(server string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	
	for i, s := range n.Servers {
		if s == server {
			// 通过切片移除服务器
			n.Servers = append(n.Servers[:i], n.Servers[i+1:]...)
			return true
		}
	}
	
	return false
}

// GetServers 返回已配置的NTP服务器列表
func (n *NTPSync) GetServers() []string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	// 返回副本以防止外部修改
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	
	return servers
}

// SetSyncInterval 设置自动同步的时间间隔
func (n *NTPSync) SetSyncInterval(interval time.Duration) {
	if interval <= 0 {
		interval = DefaultSyncInterval
	}
	
	n.mutex.Lock()
	n.SyncInterval = interval
	n.mutex.Unlock()
}

// SetTimeout 设置NTP请求的超时时间
func (n *NTPSync) SetTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	
	n.mutex.Lock()
	n.Timeout = timeout
	n.mutex.Unlock()
}

// SyncAsync 执行异步同步并立即返回
func (n *NTPSync) SyncAsync() {
	go func() {
		_ = n.Sync()
	}()
}
