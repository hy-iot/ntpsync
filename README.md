# Go NTP Sync

一个纯Go语言实现的NTP时间同步包，支持多服务器和定时同步功能。该包不依赖任何第三方库，仅使用Go标准库实现。

## 功能特点

- 纯Go语言实现，无第三方依赖
- 支持多个NTP服务器配置
- 支持服务器自动故障转移
- 支持并行查询多个NTP服务器
- 支持定时自动同步
- 支持服务器状态监控
- 线程安全设计
- 精确的时间偏移计算
- 简单易用的API

## 安装

```bash
go get github.com/hy-iot/ntpsync
```

## 快速开始

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/hy-iot/ntpsync/pkg/ntpsync"
)

func main() {
    // 创建NTP同步客户端
    ntp, err := ntpsync.New(ntpsync.Options{
        Servers: []string{
            "pool.ntp.org",
            "time.google.com",
        },
        Timeout: 5 * time.Second,
        SyncInterval: 1 * time.Hour,
        AutoSync: true, // 启用自动同步
        EnableMultiServer: true, // 启用多服务器支持
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
    
    // 获取服务器状态
    statuses, err := ntp.GetMultiServerStatus()
    if err == nil {
        fmt.Println("服务器状态:")
        for _, status := range statuses {
            fmt.Printf("  %s: 可达=%v, 层级=%d, RTT=%v\n",
                status.Address, status.Reachable, status.Stratum, status.RTT)
        }
    }
    
    // 程序结束前停止定时同步
    defer ntp.StopPeriodicSync()
}
```

## 详细使用说明

### 基本用法

创建NTP同步客户端并执行一次同步：

```go
// 创建NTP同步客户端
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{"pool.ntp.org"},
})

if err != nil {
    // 处理错误
}

// 执行一次同步
err = ntp.Sync()
if err != nil {
    // 处理错误
}

// 获取NTP校准后的当前时间
ntpTime := ntp.Now()
```

### 多服务器支持

配置多个NTP服务器以提高可靠性：

```go
// 创建支持多服务器的NTP同步客户端
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{
        "pool.ntp.org",
        "time.google.com",
        "time.windows.com",
        "time.apple.com",
    },
    EnableMultiServer: true, // 启用多服务器支持
})

// 获取所有服务器状态
statuses, err := ntp.GetMultiServerStatus()
if err != nil {
    // 处理错误
}

// 查看每个服务器的状态
for _, status := range statuses {
    fmt.Printf("服务器: %s, 可达性: %v, 层级: %d, RTT: %v\n",
        status.Address, status.Reachable, status.Stratum, status.RTT)
}

// 获取最佳服务器
bestServer, err := ntp.GetBestServer()
if err != nil {
    // 处理错误
}
fmt.Printf("最佳服务器: %s\n", bestServer)
```

### 定时同步

启用定时自动同步功能：

```go
// 创建带有定时同步的NTP客户端
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{"pool.ntp.org"},
    SyncInterval: 1 * time.Hour, // 每小时同步一次
    AutoSync: true, // 启用自动同步
})

// 或者手动启动定时同步
if !ntp.IsPeriodicSyncRunning() {
    err = ntp.StartPeriodicSync()
    if err != nil {
        // 处理错误
    }
}

// 获取定时同步状态
status := ntp.GetPeriodicSyncStatus()
fmt.Printf("同步状态: 运行=%v, 上次同步=%v, 间隔=%v\n",
    status.Running, status.LastSync, status.Interval)

// 停止定时同步
ntp.StopPeriodicSync()
```

## API参考

### 类型

#### `Options`

创建NTP同步客户端的选项：

```go
type Options struct {
    // NTP服务器地址列表
    Servers []string
    
    // NTP请求超时时间
    Timeout time.Duration
    
    // 自动同步的时间间隔
    SyncInterval time.Duration
    
    // 是否启用自动同步
    AutoSync bool
    
    // 是否启用多服务器支持
    EnableMultiServer bool
}
```

#### `ServerStatus`

NTP服务器状态：

```go
type ServerStatus struct {
    // 服务器地址
    Address string
    
    // 服务器是否可达
    Reachable bool
    
    // 最后一次成功响应的时间
    LastResponse time.Time
    
    // 往返时间
    RTT time.Duration
    
    // 服务器层级
    Stratum uint8
    
    // 时间偏移量
    Offset time.Duration
}
```

### 主要方法

- `New(opts Options) (*NTPSync, error)` - 创建新的NTP同步客户端
- `Sync() error` - 执行一次同步
- `Now() time.Time` - 获取校准后的当前时间
- `TimeOffsetDuration() time.Duration` - 获取时间偏移量
- `AddServer(server string)` - 添加NTP服务器
- `RemoveServer(server string) bool` - 移除NTP服务器
- `GetServers() []string` - 获取服务器列表
- `StartPeriodicSync() error` - 启动定时同步
- `StopPeriodicSync()` - 停止定时同步
- `GetMultiServerStatus() ([]ServerStatus, error)` - 获取所有服务器状态
- `GetBestServer() (string, error)` - 获取最佳服务器

更多详细API说明请参考[USAGE.md](USAGE.md)文档。

## 示例代码

完整的示例代码可以在[example/main.go](example/main.go)中找到。

## 许可证

MIT
