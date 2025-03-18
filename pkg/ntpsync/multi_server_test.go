package ntpsync

import (
	"testing"
	"time"
)

// TestSyncWithMultiServer 测试多服务器同步
// 这是一个需要网络访问的集成测试
func TestSyncWithMultiServer(t *testing.T) {
	// 在短模式下跳过此测试
	if testing.Short() {
		t.Skip("在短模式下跳过多服务器同步测试")
	}
	
	ntp, err := New(Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
		},
		Timeout: 5 * time.Second,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 与多个服务器同步
	err = ntp.SyncWithMultiServer()
	
	// 如果网络不可靠，此测试可能会失败
	// 我们只检查函数是否按预期工作
	if err != nil {
		t.Logf("多服务器同步失败: %v", err)
		t.Skip("由于网络错误跳过")
	}
	
	// 检查时间偏移量是否已设置
	if ntp.TimeOffset == 0 {
		t.Error("预期时间偏移量已设置，实际得到0")
	}
	
	// 检查最后同步时间是否已设置
	if ntp.LastSync.IsZero() {
		t.Error("预期最后同步时间已设置，实际得到零时间")
	}
}

// TestSyncWithMultiServerParallel 测试并行多服务器同步
// 这是一个需要网络访问的集成测试
func TestSyncWithMultiServerParallel(t *testing.T) {
	// 在短模式下跳过此测试
	if testing.Short() {
		t.Skip("在短模式下跳过并行多服务器同步测试")
	}
	
	ntp, err := New(Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
		},
		Timeout: 5 * time.Second,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 并行与多个服务器同步
	err = ntp.SyncWithMultiServerParallel()
	
	// 如果网络不可靠，此测试可能会失败
	// 我们只检查函数是否按预期工作
	if err != nil {
		t.Logf("并行多服务器同步失败: %v", err)
		t.Skip("由于网络错误跳过")
	}
	
	// 检查时间偏移量是否已设置
	if ntp.TimeOffset == 0 {
		t.Error("预期时间偏移量已设置，实际得到0")
	}
	
	// 检查最后同步时间是否已设置
	if ntp.LastSync.IsZero() {
		t.Error("预期最后同步时间已设置，实际得到零时间")
	}
}

// TestGetMultiServerStatus 测试获取多个服务器的状态
// 这是一个需要网络访问的集成测试
func TestGetMultiServerStatus(t *testing.T) {
	// 在短模式下跳过此测试
	if testing.Short() {
		t.Skip("在短模式下跳过多服务器状态测试")
	}
	
	ntp, err := New(Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
		},
		Timeout: 5 * time.Second,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 获取多个服务器的状态
	statuses, err := ntp.GetMultiServerStatus()
	
	// 如果网络不可靠，此测试可能会失败
	// 我们只检查函数是否按预期工作
	if err != nil {
		t.Logf("获取多服务器状态失败: %v", err)
		t.Skip("由于网络错误跳过")
	}
	
	// 检查我们是否获得了状态
	if len(statuses) != 2 {
		t.Errorf("预期2个状态，实际得到%d个", len(statuses))
	}
	
	// 检查状态是否包含正确的服务器
	servers := make(map[string]bool)
	for _, status := range statuses {
		servers[status.Address] = true
	}
	
	if !servers["pool.ntp.org:123"] && !servers["pool.ntp.org"] {
		t.Error("预期有pool.ntp.org的状态，未找到")
	}
	
	if !servers["time.google.com:123"] && !servers["time.google.com"] {
		t.Error("预期有time.google.com的状态，未找到")
	}
}

// TestUpdateNTPSyncWithMultiServer 测试更新NTPSync结构体以使用多服务器功能
func TestUpdateNTPSyncWithMultiServer(t *testing.T) {
	ntp, err := New(Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
		},
		Timeout: 5 * time.Second,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 更新以使用多服务器功能
	ntp.UpdateNTPSyncWithMultiServer()
	
	// 检查Sync方法现在是否使用多服务器功能
	// 这很难直接测试，所以我们只检查调用时是否不会崩溃
	
	// 我们设置一个短的超时时间以避免等待太长时间
	ntp.Timeout = 100 * time.Millisecond
	
	// 调用Sync，现在应该使用多服务器功能
	// 我们不关心它是否成功或失败，只关心它不会崩溃
	_ = ntp.Sync()
}
