// Package ntpsync 提供NTP时间同步功能。
package ntpsync

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

// UpdateSystemTime 使用NTP同步的时间更新系统时间
// 注意：此操作通常需要root/管理员权限
func (n *NTPSync) UpdateSystemTime() error {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	// 首先确保我们有有效的时间偏移量
	if n.LastSync.IsZero() {
		// 尝试同步
		if err := n.Sync(); err != nil {
			return fmt.Errorf("无法同步NTP时间: %w", err)
		}
	}

	// 获取当前NTP调整后的时间
	ntpTime := n.Now()

	// 根据操作系统设置系统时间
	switch runtime.GOOS {
	case "linux", "darwin":
		// 使用date命令设置时间 (需要root权限)
		// 格式: MMDDhhmm[[CC]YY][.ss]
		timeStr := ntpTime.Format("010215042006.05")
		cmd := exec.Command("date", timeStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("设置系统时间失败: %w, 输出: %s", err, output)
		}

	case "windows":
		// 使用PowerShell设置时间 (需要管理员权限)
		dateStr := ntpTime.Format("01/02/2006")
		timeStr := ntpTime.Format("15:04:05")
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Set-Date -Date '%s %s'", dateStr, timeStr))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("设置系统时间失败: %w, 输出: %s", err, output)
		}

	default:
		return errors.New("不支持的操作系统")
	}

	return nil
}

// IsRootUser 检查当前进程是否具有root/管理员权限
// 这个函数可以用来在尝试更新系统时间前检查权限
func IsRootUser() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		// 在Unix系统上，尝试运行一个需要root权限的简单命令
		cmd := exec.Command("id", "-u")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		// root用户的ID是0
		return string(output) == "0\n"

	case "windows":
		// 在Windows上检查是否有管理员权限
		cmd := exec.Command("powershell", "-Command",
			"[bool](([System.Security.Principal.WindowsIdentity]::GetCurrent()).groups -match 'S-1-5-32-544')")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		return string(output) == "True\n"

	default:
		return false
	}
}
