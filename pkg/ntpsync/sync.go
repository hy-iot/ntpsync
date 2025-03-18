package ntpsync

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// syncWithServer 与特定的NTP服务器同步
func (n *NTPSync) syncWithServer(server string, timeout time.Duration) (*SyncResult, error) {
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

	// 创建并发送NTP请求数据包
	req := createNTPPacket()
	t1 := time.Now() // 发送请求的时间
	
	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return nil, fmt.Errorf("发送NTP请求失败: %v", err)
	}

	// 接收响应
	resp := &NTPPacket{}
	if err := binary.Read(conn, binary.BigEndian, resp); err != nil {
		return nil, fmt.Errorf("读取NTP响应失败: %v", err)
	}
	
	t4 := time.Now() // 接收响应的时间

	// 验证响应
	if resp.Stratum == 0 {
		return nil, errors.New("服务器返回无效的0层级响应")
	}

	// 计算时间
	t2 := ntpTimeToTime(resp.RxTimeSec, resp.RxTimeFrac)
	t3 := ntpTimeToTime(resp.TxTimeSec, resp.TxTimeFrac)

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
		Stratum: resp.Stratum,
	}

	return result, nil
}

// createNTPPacket 创建一个用于客户端请求的NTP数据包
func createNTPPacket() *NTPPacket {
	packet := &NTPPacket{
		// LI (0), VN (4), Mode (3)
		Settings: uint8((0 << 6) | (4 << 3) | (3)),
	}
	
	// 设置发送时间戳为当前时间
	now := time.Now()
	seconds, fraction := timeToNTPTime(now)
	packet.TxTimeSec = seconds
	packet.TxTimeFrac = fraction
	
	return packet
}
