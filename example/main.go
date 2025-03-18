package main

import (
	"fmt"
	"time"

	"github.com/hy-iot/ntpsync/pkg/ntpsync"
)

func main() {
	// 创建一个支持多服务器的NTP同步客户端
	ntp, err := ntpsync.New(ntpsync.Options{
		Servers: []string{
			"pool.ntp.org",
			"time.google.com",
			"time.windows.com",
			"time.apple.com",
		},
		Timeout:           5 * time.Second,
		SyncInterval:      1 * time.Minute, // 演示用的短间隔
		AutoSync:          true,            // 启用自动同步
		EnableMultiServer: true,            // 启用多服务器支持
	})

	if err != nil {
		fmt.Printf("创建NTP客户端失败: %v\n", err)
		return
	}

	// 获取NTP校准后的当前时间
	ntpTime := ntp.Now()
	fmt.Printf("当前NTP时间: %v\n", ntpTime)

	// 获取时间偏移量
	offset := ntp.TimeOffsetDuration()
	fmt.Printf("时间偏移量: %v\n", offset)

	// 获取已配置的服务器列表
	servers := ntp.GetServers()
	fmt.Printf("已配置的服务器: %v\n", servers)

	// 等待几秒钟以允许初始同步
	fmt.Println("等待初始同步...")
	time.Sleep(5 * time.Second)

	// 获取所有服务器的状态
	statuses, err := ntp.GetMultiServerStatus()
	if err != nil {
		fmt.Printf("获取服务器状态失败: %v\n", err)
	} else {
		fmt.Println("服务器状态:")
		for _, status := range statuses {
			fmt.Printf("  %s: 可达=%v, 层级=%d, RTT=%v\n",
				status.Address, status.Reachable, status.Stratum, status.RTT)
		}
	}

	// 找出最佳服务器（层级最低且RTT最小的服务器）
	var bestServer string
	var bestStratum uint8 = 16 // NTP最大层级为16
	var bestRTT time.Duration = time.Hour

	for _, status := range statuses {
		if status.Reachable && (status.Stratum < bestStratum ||
			(status.Stratum == bestStratum && status.RTT < bestRTT)) {
			bestServer = status.Address
			bestStratum = status.Stratum
			bestRTT = status.RTT
		}
	}

	if bestServer != "" {
		fmt.Printf("最佳服务器: %s (层级=%d, RTT=%v)\n",
			bestServer, bestStratum, bestRTT)
	} else {
		fmt.Println("未找到可用的服务器")
	}

	// 获取定时同步状态
	syncStatus := ntp.GetPeriodicSyncStatus()
	fmt.Printf("定时同步状态: 运行中=%v, 上次同步=%v, 间隔=%v\n",
		syncStatus.Running, syncStatus.LastSync, syncStatus.Interval)

	// 强制同步
	fmt.Println("强制同步...")
	err = ntp.ForceSyncNow()
	if err != nil {
		fmt.Printf("强制同步失败: %v\n", err)
	} else {
		fmt.Println("强制同步成功")
	}

	// 获取更新后的时间和偏移量
	ntpTime = ntp.Now()
	offset = ntp.TimeOffsetDuration()
	fmt.Printf("更新后的NTP时间: %v\n", ntpTime)
	fmt.Printf("更新后的时间偏移量: %v\n", offset)

	// 检查是否有root/管理员权限
	isRoot := ntpsync.IsRootUser()
	fmt.Printf("是否有root/管理员权限: %v\n", isRoot)

	// 如果有root权限，尝试更新系统时间
	if isRoot {
		fmt.Println("尝试更新系统时间...")
		err = ntp.UpdateSystemTime()
		if err != nil {
			fmt.Printf("更新系统时间失败: %v\n", err)
		} else {
			fmt.Println("系统时间已成功更新")
		}
	} else {
		fmt.Println("没有root/管理员权限，无法更新系统时间")
	}

	// 添加新服务器
	fmt.Println("添加新服务器...")
	ntp.AddServer("time.cloudflare.com")
	servers = ntp.GetServers()
	fmt.Printf("更新后的服务器: %v\n", servers)

	// 移除服务器
	fmt.Println("移除服务器...")
	ntp.RemoveServer("time.apple.com")
	servers = ntp.GetServers()
	fmt.Printf("更新后的服务器: %v\n", servers)

	// 更改同步间隔
	fmt.Println("更改同步间隔...")
	ntp.SetPeriodicSyncInterval(30 * time.Second)
	interval := ntp.GetPeriodicSyncInterval()
	fmt.Printf("新的同步间隔: %v\n", interval)

	// 等待一段时间以观察定时同步
	fmt.Println("等待观察定时同步...")
	for i := 0; i < 3; i++ {
		time.Sleep(10 * time.Second)
		syncStatus = ntp.GetPeriodicSyncStatus()
		fmt.Printf("同步状态（%d0秒后）: 成功次数=%d, 错误次数=%d, 上次同步=%v\n",
			i+1, syncStatus.SuccessCount, syncStatus.ErrorCount, syncStatus.LastSync)
	}

	// 停止同步
	fmt.Println("停止同步...")
	ntp.StopPeriodicSync()

	// 检查是否已停止
	syncStatus = ntp.GetPeriodicSyncStatus()
	fmt.Printf("停止后的同步状态: 运行中=%v\n", syncStatus.Running)

	fmt.Println("示例成功完成")
}
