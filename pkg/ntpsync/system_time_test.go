package ntpsync

import (
	"testing"
	"time"
)

// TestIsRootUser 测试权限检查功能
func TestIsRootUser(t *testing.T) {
	// 这个测试只检查函数是否能正常运行，不检查实际结果
	// 因为测试环境可能有不同的权限
	isRoot := IsRootUser()

	// 记录结果，但不断言特定值
	t.Logf("当前进程是否有root/管理员权限: %v", isRoot)
}

// TestUpdateSystemTime 测试更新系统时间功能
// 注意：这个测试不会实际更新系统时间，只检查函数逻辑
func TestUpdateSystemTime(t *testing.T) {
	// 跳过实际测试，因为它需要root权限
	t.Skip("跳过系统时间更新测试，因为它需要root/管理员权限")

	// 以下代码仅用于说明，不会实际执行

	// 创建一个NTP客户端
	ntp, err := New(Options{
		Servers: []string{"pool.ntp.org"},
	})

	if err != nil {
		t.Fatalf("创建NTP客户端失败: %v", err)
	}

	// 设置一个已知的偏移量
	ntp.TimeOffset = 5 * time.Second
	ntp.LastSync = time.Now()

	// 检查是否有root权限
	isRoot := IsRootUser()
	t.Logf("当前进程是否有root/管理员权限: %v", isRoot)

	// 如果有root权限，可以尝试更新系统时间
	if isRoot {
		err = ntp.UpdateSystemTime()
		if err != nil {
			t.Fatalf("更新系统时间失败: %v", err)
		}
		t.Log("系统时间已成功更新")
	} else {
		t.Log("没有root/管理员权限，无法更新系统时间")
	}
}
