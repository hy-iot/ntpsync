package ntpsync

import (
	"errors"
	"sync/atomic"
	"time"
)

// PeriodicSyncStatus 表示定时同步的状态
type PeriodicSyncStatus struct {
	// Running 表示定时同步是否正在运行
	Running bool
	
	// LastSync 是最后一次成功同步的时间
	LastSync time.Time
	
	// LastError 是同步过程中发生的最后一个错误
	LastError error
	
	// Interval 是当前定时同步的时间间隔
	Interval time.Duration
	
	// SuccessCount 是成功同步的次数
	SuccessCount int64
	
	// ErrorCount 是失败同步的次数
	ErrorCount int64
}

// StartPeriodicSync 开始定时同步过程
func (n *NTPSync) StartPeriodicSync() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	
	// 检查是否已经在运行
	select {
	case <-n.stopChan:
		// 通道已关闭，需要重新创建
		n.stopChan = make(chan struct{})
	default:
		// 通道是开放的，这意味着同步已经在运行
		return errors.New("同步已经在运行中")
	}
	
	// 执行初始同步
	go func() {
		err := n.Sync()
		if err != nil {
			atomic.AddInt64(&n.errorCount, 1)
			n.mutex.Lock()
			n.lastError = err
			n.mutex.Unlock()
		} else {
			atomic.AddInt64(&n.successCount, 1)
			n.mutex.Lock()
			n.LastSync = time.Now()
			n.mutex.Unlock()
		}
	}()
	
	// 启动同步goroutine
	n.syncWaitGroup.Add(1)
	go n.periodicSyncLoop()
	
	n.AutoSync = true
	return nil
}

// StopPeriodicSync 停止定时同步过程
func (n *NTPSync) StopPeriodicSync() {
	n.mutex.Lock()
	
	// 检查是否已经停止
	select {
	case <-n.stopChan:
		// 已经停止
		n.mutex.Unlock()
		return
	default:
		// 关闭通道以发出停止信号
		close(n.stopChan)
	}
	
	n.AutoSync = false
	n.mutex.Unlock()
	
	// 等待同步循环退出
	n.syncWaitGroup.Wait()
}

// periodicSyncLoop 是定时同步的主循环
func (n *NTPSync) periodicSyncLoop() {
	defer n.syncWaitGroup.Done()
	
	for {
		// 获取当前同步间隔
		n.mutex.RLock()
		interval := n.SyncInterval
		n.mutex.RUnlock()
		
		// 为下一次同步创建定时器
		timer := time.NewTimer(interval)
		
		// 等待定时器或停止信号
		select {
		case <-timer.C:
			// 同步时间到
			err := n.Sync()
			if err != nil {
				atomic.AddInt64(&n.errorCount, 1)
				n.mutex.Lock()
				n.lastError = err
				n.mutex.Unlock()
			} else {
				atomic.AddInt64(&n.successCount, 1)
				n.mutex.Lock()
				n.LastSync = time.Now()
				n.mutex.Unlock()
			}
		case <-n.stopChan:
			// 请求停止
			if !timer.Stop() {
				// 如果定时器已经触发，则排空定时器通道
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}
}

// IsPeriodicSyncRunning 返回定时同步是否正在运行
func (n *NTPSync) IsPeriodicSyncRunning() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	select {
	case <-n.stopChan:
		return false
	default:
		return n.AutoSync
	}
}

// GetPeriodicSyncStatus 返回定时同步的当前状态
func (n *NTPSync) GetPeriodicSyncStatus() PeriodicSyncStatus {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	status := PeriodicSyncStatus{
		Running:      n.IsPeriodicSyncRunning(),
		LastSync:     n.LastSync,
		LastError:    n.lastError,
		Interval:     n.SyncInterval,
		SuccessCount: atomic.LoadInt64(&n.successCount),
		ErrorCount:   atomic.LoadInt64(&n.errorCount),
	}
	
	return status
}

// ForceSyncNow 强制立即同步
func (n *NTPSync) ForceSyncNow() error {
	err := n.Sync()
	
	if err != nil {
		atomic.AddInt64(&n.errorCount, 1)
		n.mutex.Lock()
		n.lastError = err
		n.mutex.Unlock()
	} else {
		atomic.AddInt64(&n.successCount, 1)
		n.mutex.Lock()
		n.LastSync = time.Now()
		n.mutex.Unlock()
	}
	
	return err
}

// SetPeriodicSyncInterval 设置定时同步的时间间隔
func (n *NTPSync) SetPeriodicSyncInterval(interval time.Duration) {
	if interval <= 0 {
		interval = DefaultSyncInterval
	}
	
	n.mutex.Lock()
	n.SyncInterval = interval
	n.mutex.Unlock()
}

// GetPeriodicSyncInterval 返回当前定时同步的时间间隔
func (n *NTPSync) GetPeriodicSyncInterval() time.Duration {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	return n.SyncInterval
}
