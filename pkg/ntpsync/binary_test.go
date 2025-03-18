package ntpsync

import (
	"testing"
	"time"
)

// TestCreateNTPPacket 测试NTP数据包的创建
func TestCreateNTPPacket(t *testing.T) {
	// 创建数据包
	packet := createNTPPacket()
	
	// 检查设置字节
	// LI (0), VN (4), Mode (3)
	expectedSettings := uint8((0 << 6) | (4 << 3) | (3))
	if packet.Settings != expectedSettings {
		t.Errorf("预期设置为 %d，实际得到 %d", expectedSettings, packet.Settings)
	}
	
	// 检查发送时间戳是否已设置
	if packet.TxTimeSec == 0 || packet.TxTimeFrac == 0 {
		t.Error("预期发送时间戳已设置，实际得到 0")
	}
}

// TestTimeToNTPTime 测试从time.Time到NTP时间的转换
func TestTimeToNTPTime(t *testing.T) {
	// 使用已知时间进行测试
	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 转换为NTP时间
	seconds, fraction := timeToNTPTime(testTime)
	
	// 预期值
	// 2020-01-01 00:00:00 UTC 是自1900-01-01以来的3786825600秒
	expectedSeconds := uint32(3786825600)
	
	// 检查秒数
	if seconds != expectedSeconds {
		t.Errorf("预期秒数为 %d，实际得到 %d", expectedSeconds, seconds)
	}
	
	// 检查小数部分（对于没有纳秒的时间应为0）
	if fraction != 0 {
		t.Errorf("预期小数部分为 0，实际得到 %d", fraction)
	}
	
	// 使用带有纳秒的时间进行测试
	testTime = time.Date(2020, 1, 1, 0, 0, 0, 500000000, time.UTC)
	
	// 转换为NTP时间
	seconds, fraction = timeToNTPTime(testTime)
	
	// 预期值
	// 500000000纳秒是0.5秒，在NTP小数部分中为0x80000000
	expectedFraction := uint32(0x80000000)
	
	// 检查小数部分
	if fraction != expectedFraction {
		t.Errorf("预期小数部分为 %d，实际得到 %d", expectedFraction, fraction)
	}
}

// TestNTPTimeToTime 测试从NTP时间到time.Time的转换
func TestNTPTimeToTime(t *testing.T) {
	// 使用已知的NTP值进行测试
	seconds := uint32(3786825600) // 2020-01-01 00:00:00 UTC
	fraction := uint32(0x80000000) // 0.5秒
	
	// 转换为time.Time
	tm := ntpTimeToTime(seconds, fraction)
	
	// 预期时间
	expectedTime := time.Date(2020, 1, 1, 0, 0, 0, 500000000, time.UTC)
	
	// 检查时间
	if !tm.Equal(expectedTime) {
		t.Errorf("预期时间为 %v，实际得到 %v", expectedTime, tm)
	}
}

// TestSyncWithServerBinary 测试与服务器的二进制同步
// 这是一个需要网络访问的集成测试
func TestSyncWithServerBinary(t *testing.T) {
	// 在短模式下跳过此测试
	if testing.Short() {
		t.Skip("在短模式下跳过服务器同步测试")
	}
	
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
		Timeout: 5 * time.Second,
	})
	
	if err != nil {
		t.Fatalf("创建NTPSync实例失败: %v", err)
	}
	
	// 与服务器同步
	result, err := ntp.syncWithServerBinary("pool.ntp.org", 5*time.Second)
	
	// 如果网络不可靠，此测试可能会失败
	// 我们只检查函数是否按预期工作
	if err != nil {
		t.Logf("同步失败: %v", err)
		t.Skip("由于网络错误跳过")
	}
	
	// 检查结果
	if result == nil {
		t.Fatal("预期有结果，实际得到nil")
	}
	
	// 检查服务器是否已设置
	if result.Server != "pool.ntp.org:123" {
		t.Errorf("预期服务器为pool.ntp.org:123，实际得到 %s", result.Server)
	}
	
	// 检查时间是否已设置
	if result.Time.IsZero() {
		t.Error("预期时间已设置，实际得到零时间")
	}
	
	// 检查偏移量是否合理
	// 偏移量应在几秒钟之内
	if result.Offset < -10*time.Second || result.Offset > 10*time.Second {
		t.Errorf("预期偏移量合理，实际得到 %v", result.Offset)
	}
	
	// 检查RTT是否合理
	// RTT应为正值且小于超时时间
	if result.RTT <= 0 || result.RTT >= 5*time.Second {
		t.Errorf("预期RTT合理，实际得到 %v", result.RTT)
	}
	
	// 检查层级是否合理
	// 层级应在1到15之间
	if result.Stratum < 1 || result.Stratum > 15 {
		t.Errorf("预期层级合理，实际得到 %d", result.Stratum)
	}
}
