package ntpsync

import (
	"testing"
	"time"
)

// TestNew 测试创建新的NTPSync实例
func TestNew(t *testing.T) {
	// 使用有效选项进行测试
	opts := Options{
		Servers: []string{"pool.ntp.org"},
		Timeout: 5 * time.Second,
	}
	
	ntp, err := New(opts)
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	if ntp == nil {
		t.Fatal("NTPSync实例为nil")
	}
	
	if len(ntp.Servers) != 1 {
		t.Errorf("预期1个服务器，实际得到%d个", len(ntp.Servers))
	}
	
	if ntp.Timeout != 5*time.Second {
		t.Errorf("预期超时时间为5秒，实际得到%v", ntp.Timeout)
	}
	
	// 测试没有服务器的情况
	opts = Options{
		Servers: []string{},
	}
	
	_, err = New(opts)
	if err == nil {
		t.Error("预期没有服务器时会返回错误，实际得到nil")
	}
	
	// 测试默认值
	opts = Options{
		Servers: []string{"pool.ntp.org"},
	}
	
	ntp, err = New(opts)
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	if ntp.Timeout != DefaultTimeout {
		t.Errorf("预期默认超时时间，实际得到%v", ntp.Timeout)
	}
	
	if ntp.SyncInterval != DefaultSyncInterval {
		t.Errorf("预期默认同步间隔，实际得到%v", ntp.SyncInterval)
	}
}

// TestAddRemoveServer 测试添加和移除服务器
func TestAddRemoveServer(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 测试添加服务器
	ntp.AddServer("time.google.com")
	
	servers := ntp.GetServers()
	if len(servers) != 2 {
		t.Errorf("预期2个服务器，实际得到%d个", len(servers))
	}
	
	// 测试添加重复的服务器
	ntp.AddServer("time.google.com")
	
	servers = ntp.GetServers()
	if len(servers) != 2 {
		t.Errorf("预期2个服务器（无重复），实际得到%d个", len(servers))
	}
	
	// 测试移除服务器
	removed := ntp.RemoveServer("time.google.com")
	if !removed {
		t.Error("预期服务器被移除，实际得到false")
	}
	
	servers = ntp.GetServers()
	if len(servers) != 1 {
		t.Errorf("预期移除后剩余1个服务器，实际得到%d个", len(servers))
	}
	
	// 测试移除不存在的服务器
	removed = ntp.RemoveServer("non.existent.server")
	if removed {
		t.Error("预期对不存在的服务器返回false，实际得到true")
	}
}

// TestSetInterval 测试设置同步间隔
func TestSetInterval(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 测试设置有效的间隔
	ntp.SetSyncInterval(10 * time.Minute)
	
	if ntp.SyncInterval != 10*time.Minute {
		t.Errorf("预期间隔为10分钟，实际得到%v", ntp.SyncInterval)
	}
	
	// 测试设置无效的间隔
	ntp.SetSyncInterval(-5 * time.Second)
	
	if ntp.SyncInterval != DefaultSyncInterval {
		t.Errorf("预期无效值时使用默认间隔，实际得到%v", ntp.SyncInterval)
	}
}

// TestSetTimeout 测试设置超时时间
func TestSetTimeout(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 测试设置有效的超时时间
	ntp.SetTimeout(10 * time.Second)
	
	if ntp.Timeout != 10*time.Second {
		t.Errorf("预期超时时间为10秒，实际得到%v", ntp.Timeout)
	}
	
	// 测试设置无效的超时时间
	ntp.SetTimeout(-5 * time.Second)
	
	if ntp.Timeout != DefaultTimeout {
		t.Errorf("预期无效值时使用默认超时时间，实际得到%v", ntp.Timeout)
	}
}

// TestNow 测试Now方法
func TestNow(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 设置已知的偏移量
	offset := 5 * time.Second
	ntp.TimeOffset = offset
	
	// 获取时间
	ntpTime := ntp.Now()
	systemTime := time.Now()
	
	// 检查偏移量是否已应用
	diff := ntpTime.Sub(systemTime)
	
	// 允许由于执行时间导致的小误差
	if diff < offset-100*time.Millisecond || diff > offset+100*time.Millisecond {
		t.Errorf("预期时间差异约为%v，实际得到%v", offset, diff)
	}
}

// TestMultiServerSupport 测试多服务器支持
func TestMultiServerSupport(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
		},
		EnableMultiServer: true,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 检查服务器管理器是否已创建
	if ntp.serverManager == nil {
		t.Fatal("预期服务器管理器已创建，实际得到nil")
	}
	
	// 检查服务器是否已添加到管理器
	servers := ntp.serverManager.GetServers()
	if len(servers) != 2 {
		t.Errorf("预期管理器中有2个服务器，实际得到%d个", len(servers))
	}
}

// TestPeriodicSync 测试定时同步功能
func TestPeriodicSync(t *testing.T) {
	// 在短模式下跳过此测试
	if testing.Short() {
		t.Skip("在短模式下跳过定时同步测试")
	}
	
	// 创建一个没有自动同步的NTP客户端
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
		SyncInterval: 1 * time.Second, // 用于测试的短间隔
		AutoSync: false, // 确保不自动启动同步
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 确保stopChan处于关闭状态
	ntp.StopPeriodicSync()
	
	// 启动定时同步
	err = ntp.StartPeriodicSync()
	if err != nil {
		t.Fatalf("启动定时同步失败: %v", err)
	}
	
	// 检查是否正在运行
	if !ntp.IsPeriodicSyncRunning() {
		t.Error("预期定时同步正在运行，实际得到false")
	}
	
	// 等待几个同步周期
	time.Sleep(3 * time.Second)
	
	// 停止定时同步
	ntp.StopPeriodicSync()
	
	// 检查是否已停止
	if ntp.IsPeriodicSyncRunning() {
		t.Error("预期定时同步已停止，实际得到true")
	}
	
	// 检查同步状态
	status := ntp.GetPeriodicSyncStatus()
	
	// 应该至少有一次同步尝试
	totalCount := status.SuccessCount + status.ErrorCount
	if totalCount < 1 {
		t.Errorf("预期至少有1次同步尝试，实际得到%d次", totalCount)
	}
}
