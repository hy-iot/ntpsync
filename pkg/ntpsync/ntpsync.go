// Package ntpsync 提供NTP时间同步功能。
// 支持多个NTP服务器和定时同步。
package ntpsync

import (
	"errors"
	"sync"
	"time"
)

// NTP协议相关常量
const (
	// DefaultNTPPort 是标准NTP端口
	DefaultNTPPort = "123"
	
	// DefaultTimeout 是NTP请求的默认超时时间
	DefaultTimeout = 5 * time.Second
	
	// DefaultSyncInterval 是定时同步的默认间隔
	DefaultSyncInterval = 1 * time.Hour
)

// NTPSync 表示一个NTP同步客户端
type NTPSync struct {
	// Servers 是NTP服务器地址列表
	Servers []string
	
	// Timeout 是NTP请求的超时时间
	Timeout time.Duration
	
	// SyncInterval 是自动同步的时间间隔
	SyncInterval time.Duration
	
	// TimeOffset 是本地时间与NTP时间的计算偏移量
	TimeOffset time.Duration
	
	// LastSync 是最后一次成功同步的时间
	LastSync time.Time
	
	// AutoSync 表示是否启用自动同步
	AutoSync bool
	
	// stopChan 用于停止自动同步
	stopChan chan struct{}
	
	// mutex 用于线程安全
	mutex sync.RWMutex
	
	// syncWaitGroup 用于优雅关闭
	syncWaitGroup sync.WaitGroup
	
	// serverManager 处理多个NTP服务器和故障转移
	serverManager *ServerManager
	
	// lastError 是同步过程中发生的最后一个错误
	lastError error
	
	// successCount 是成功同步的次数
	successCount int64
	
	// errorCount 是失败同步的次数
	errorCount int64
}

// Options 包含NTPSync的配置选项
type Options struct {
	// Servers 是NTP服务器地址列表
	Servers []string
	
	// Timeout 是NTP请求的超时时间
	Timeout time.Duration
	
	// SyncInterval 是自动同步的时间间隔
	SyncInterval time.Duration
	
	// AutoSync 表示是否启用自动同步
	AutoSync bool
	
	// EnableMultiServer 表示是否启用多服务器支持
	EnableMultiServer bool
}

// New 创建一个新的NTPSync实例
func New(opts Options) (*NTPSync, error) {
	if len(opts.Servers) == 0 {
		return nil, errors.New("必须提供至少一个NTP服务器")
	}
	
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	
	syncInterval := opts.SyncInterval
	if syncInterval <= 0 {
		syncInterval = DefaultSyncInterval
	}
	
	ntp := &NTPSync{
		Servers:      opts.Servers,
		Timeout:      timeout,
		SyncInterval: syncInterval,
		AutoSync:     opts.AutoSync,
		stopChan:     make(chan struct{}),
	}
	
	// 如果启用了多服务器支持，则初始化服务器管理器
	if opts.EnableMultiServer {
		var err error
		ntp.serverManager, err = NewServerManager(opts.Servers, timeout)
		if err != nil {
			return nil, err
		}
	}
	
	// 如果启用了自动同步，则启动定时同步
	if opts.AutoSync {
		if err := ntp.StartPeriodicSync(); err != nil {
			return nil, err
		}
	}
	
	return ntp, nil
}
