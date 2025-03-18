package ntpsync

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// SyncWithBinary 使用二进制操作执行一次与NTP服务器的同步
// 此实现不依赖任何第三方包
func (n *NTPSync) SyncWithBinary() error {
	n.mutex.Lock()
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	timeout := n.Timeout
	n.mutex.Unlock()

	if len(servers) == 0 {
		return errors.New("未配置NTP服务器")
	}

	var lastErr error
	for _, server := range servers {
		result, err := n.syncWithServerBinary(server, timeout)
		if err != nil {
			lastErr = err
			continue
		}

		// 成功与此服务器同步
		n.mutex.Lock()
		n.TimeOffset = result.Offset
		n.LastSync = time.Now()
		n.mutex.Unlock()
		
		return nil
	}

	// 如果执行到这里，说明所有服务器都失败了
	return fmt.Errorf("无法与任何NTP服务器同步: %v", lastErr)
}

// syncWithServerBinary 使用直接二进制操作与特定的NTP服务器同步
func (n *NTPSync) syncWithServerBinary(server string, timeout time.Duration) (*SyncResult, error) {
	// 确保服务器地址包含端口
	if _, _, err := net.SplitHostPort(server); err != nil {
		server = net.JoinHostPort(server, DefaultNTPPort)
	}

	// 创建UDP连接
	conn, err := net.DialTimeout("udp", server, timeout)
	if err != nil {
		return nil, fmt.Errorf("连接NTP服务器 %s 失败: %v", server, err)
	}
	defer conn.Close()

	// 设置读写超时
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("设置超时时间失败: %v", err)
	}

	// 创建NTP请求数据包
	reqBytes := make([]byte, 48)
	
	// LI (0), VN (4), Mode (3)
	reqBytes[0] = (0 << 6) | (4 << 3) | (3)
	
	// 设置发送时间戳为当前时间
	t1 := time.Now() // 发送请求的时间
	seconds, fraction := timeToNTPTime(t1)
	
	// 写入发送时间戳（秒和小数部分）
	binary.BigEndian.PutUint32(reqBytes[40:], seconds)
	binary.BigEndian.PutUint32(reqBytes[44:], fraction)
	
	// 发送请求
	if _, err := conn.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("发送NTP请求失败: %v", err)
	}

	// 接收响应
	respBytes := make([]byte, 48)
	bytesRead, err := conn.Read(respBytes)
	if err != nil {
		return nil, fmt.Errorf("读取NTP响应失败: %v", err)
	}
	
	if bytesRead != 48 {
		return nil, fmt.Errorf("无效的NTP响应大小: %d", bytesRead)
	}
	
	t4 := time.Now() // 接收响应的时间

	// 解析响应
	stratum := respBytes[1]
	if stratum == 0 {
		return nil, errors.New("服务器返回无效的0层级响应")
	}

	// 提取时间戳
	rxSeconds := binary.BigEndian.Uint32(respBytes[32:36])
	rxFraction := binary.BigEndian.Uint32(respBytes[36:40])
	txSeconds := binary.BigEndian.Uint32(respBytes[40:44])
	txFraction := binary.BigEndian.Uint32(respBytes[44:48])

	// 转换为time.Time
	t2 := ntpTimeToTime(rxSeconds, rxFraction)
	t3 := ntpTimeToTime(txSeconds, txFraction)

	// 计算偏移量和往返延迟
	// 偏移量 = ((T2 - T1) + (T3 - T4)) / 2
	// 延迟 = (T4 - T1) - (T3 - T2)
	offset := ((t2.Sub(t1) + t3.Sub(t4)) / 2)
	rtt := (t4.Sub(t1) - (t3.Sub(t2)))

	if rtt < 0 {
		// 这种情况在正常操作中不应该发生
		return nil, errors.New("往返时间为负值，可能在同步过程中发生了时钟调整")
	}

	result := &SyncResult{
		Server:  server,
		Time:    time.Now().Add(offset),
		Offset:  offset,
		RTT:     rtt,
		Stratum: stratum,
	}

	return result, nil
}

// GetStatusBinary 使用二进制操作返回所有已配置NTP服务器的状态
func (n *NTPSync) GetStatusBinary() ([]ServerStatus, error) {
	n.mutex.RLock()
	servers := make([]string, len(n.Servers))
	copy(servers, n.Servers)
	timeout := n.Timeout
	n.mutex.RUnlock()

	if len(servers) == 0 {
		return nil, errors.New("未配置NTP服务器")
	}

	statuses := make([]ServerStatus, 0, len(servers))
	
	for _, server := range servers {
		status := ServerStatus{
			Address: server,
		}
		
		result, err := n.syncWithServerBinary(server, timeout)
		if err != nil {
			status.Reachable = false
		} else {
			status.Reachable = true
			status.LastResponse = time.Now()
			status.RTT = result.RTT
			status.Stratum = result.Stratum
			status.Offset = result.Offset
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses, nil
}
