package ntpsync

import (
	"time"
)

// NTPPacket 表示符合RFC 5905的NTP数据包结构
type NTPPacket struct {
	// 闰秒指示器(2位)，版本号(3位)和模式(3位)
	Settings       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

// NTPMode 表示NTP数据包的模式
type NTPMode uint8

// RFC 5905中定义的NTP模式
const (
	Reserved NTPMode = 0        // 保留
	SymActive NTPMode = 1       // 对称主动模式
	SymPassive NTPMode = 2      // 对称被动模式
	Client NTPMode = 3          // 客户端模式
	Server NTPMode = 4          // 服务器模式
	Broadcast NTPMode = 5       // 广播模式
	ControlMessage NTPMode = 6  // 控制消息
	ReservedPrivate NTPMode = 7 // 保留(私有使用)
)

// NTPLeap 表示NTP数据包的闰秒指示器
type NTPLeap uint8

// RFC 5905中定义的NTP闰秒指示器
const (
	NoWarning NTPLeap = 0       // 无警告
	LastMinute61 NTPLeap = 1    // 最后一分钟有61秒
	LastMinute59 NTPLeap = 2    // 最后一分钟有59秒
	Unsynchronized NTPLeap = 3  // 未同步
)

// NTPVersion 表示NTP协议的版本
type NTPVersion uint8

// NTP版本
const (
	Version1 NTPVersion = 1
	Version2 NTPVersion = 2
	Version3 NTPVersion = 3
	Version4 NTPVersion = 4
)

// SyncResult 表示NTP同步的结果
type SyncResult struct {
	// Server 是用于同步的NTP服务器
	Server string
	
	// Time 是同步后的时间
	Time time.Time
	
	// Offset 是本地时间与NTP时间之间的计算偏移量
	Offset time.Duration
	
	// RTT (往返时间) 是从服务器获取响应所需的时间
	RTT time.Duration
	
	// Stratum 是NTP服务器的层级
	Stratum uint8
	
	// Error 是同步过程中发生的任何错误
	Error error
}

// ServerStatus 表示NTP服务器的状态
type ServerStatus struct {
	// Address 是NTP服务器的地址
	Address string
	
	// Reachable 表示服务器是否可达
	Reachable bool
	
	// LastResponse 是最后一次成功响应的时间
	LastResponse time.Time
	
	// RTT 是最后测量的往返时间
	RTT time.Duration
	
	// Stratum 是NTP服务器的层级
	Stratum uint8
	
	// Offset 是最后测量的时间偏移量
	Offset time.Duration
}
