package ntpsync

import (
	"errors"
	"time"
)

// Start 开始自动同步过程
func (n *NTPSync) Start() error {
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
			// 记录错误或根据需要处理
			// 现在，我们将继续并稍后重试
		}
	}()
	
	// 启动同步goroutine
	n.syncWaitGroup.Add(1)
	go n.syncLoop()
	
	n.AutoSync = true
	return nil
}

// Stop 停止自动同步过程
func (n *NTPSync) Stop() {
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

// syncLoop 是自动同步的主循环
func (n *NTPSync) syncLoop() {
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
				// 记录错误或根据需要处理
				// 现在，我们将继续并稍后重试
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

// IsRunning 返回自动同步是否正在运行
func (n *NTPSync) IsRunning() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	select {
	case <-n.stopChan:
		return false
	default:
		return n.AutoSync
	}
}
