# Go NTP Sync 使用指南

本文档提供了Go NTP Sync包的详细使用说明，包括安装、配置和高级用法。

## 目录

1. [安装](#安装)
2. [基本用法](#基本用法)
3. [多服务器支持](#多服务器支持)
4. [定时同步](#定时同步)
5. [错误处理](#错误处理)
6. [高级用法](#高级用法)
7. [最佳实践](#最佳实践)

## 安装

使用Go模块安装：

```bash
go get github.com/hy-iot/ntpsync
```

在你的Go代码中导入：

```go
import "github.com/hy-iot/ntpsync/pkg/ntpsync"
```

## 基本用法

### 创建NTP客户端

```go
// 创建基本NTP客户端
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{"pool.ntp.org"},
})
if err != nil {
    log.Fatalf("创建NTP客户端失败: %v", err)
}
```

### 执行同步

```go
// 执行同步
err = ntp.Sync()
if err != nil {
    log.Printf("同步失败: %v", err)
}
```

### 获取校准后的时间

```go
// 获取校准后的当前时间
ntpTime := ntp.Now()
fmt.Printf("当前NTP时间: %v\n", ntpTime)

// 获取时间偏移量
offset := ntp.TimeOffsetDuration()
fmt.Printf("时间偏移量: %v\n", offset)

// 获取最后同步时间
lastSync := ntp.LastSyncTime()
fmt.Printf("最后同步时间: %v\n", lastSync)
```

## 多服务器支持

### 配置多个服务器

```go
// 创建支持多服务器的NTP客户端
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{
        "pool.ntp.org",
        "time.google.com",
        "time.windows.com",
    },
    EnableMultiServer: true, // 启用多服务器支持
})
```

### 获取服务器状态

```go
// 获取所有服务器状态
statuses, err := ntp.GetMultiServerStatus()
if err != nil {
    log.Printf("获取服务器状态失败: %v", err)
} else {
    for _, status := range statuses {
        fmt.Printf("服务器: %s\n", status.Address)
        fmt.Printf("  可达性: %v\n", status.Reachable)
        fmt.Printf("  层级: %d\n", status.Stratum)
        fmt.Printf("  RTT: %v\n", status.RTT)
        fmt.Printf("  偏移量: %v\n", status.Offset)
    }
}
```

### 获取最佳服务器

```go
// 获取最佳服务器
bestServer, err := ntp.GetBestServer()
if err != nil {
    log.Printf("获取最佳服务器失败: %v", err)
} else {
    fmt.Printf("最佳服务器: %s\n", bestServer)
}
```

### 并行同步

```go
// 使用并行多服务器同步
err = ntp.SyncWithMultiServerParallel()
if err != nil {
    log.Printf("并行同步失败: %v", err)
}
```

## 定时同步

### 启用自动同步

```go
// 创建时启用自动同步
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{"pool.ntp.org"},
    SyncInterval: 1 * time.Hour, // 每小时同步一次
    AutoSync: true, // 启用自动同步
})

// 或者手动启动定时同步
err = ntp.StartPeriodicSync()
if err != nil {
    log.Printf("启动定时同步失败: %v", err)
}
```

### 管理定时同步

```go
// 检查定时同步是否运行中
if ntp.IsPeriodicSyncRunning() {
    fmt.Println("定时同步正在运行")
}

// 获取定时同步状态
status := ntp.GetPeriodicSyncStatus()
fmt.Printf("同步状态:\n")
fmt.Printf("  运行: %v\n", status.Running)
fmt.Printf("  上次同步: %v\n", status.LastSync)
fmt.Printf("  同步间隔: %v\n", status.Interval)
fmt.Printf("  成功次数: %d\n", status.SuccessCount)
fmt.Printf("  失败次数: %d\n", status.ErrorCount)

// 修改同步间隔
ntp.SetPeriodicSyncInterval(30 * time.Minute)

// 强制立即同步
err = ntp.ForceSyncNow()
if err != nil {
    log.Printf("强制同步失败: %v", err)
}

// 停止定时同步
ntp.StopPeriodicSync()
```

## 错误处理

### 同步错误处理

```go
// 执行同步并处理错误
err = ntp.Sync()
if err != nil {
    switch {
    case errors.Is(err, ntpsync.ErrNoServers):
        log.Println("没有配置NTP服务器")
    case errors.Is(err, ntpsync.ErrAllServersFailed):
        log.Println("所有服务器同步失败")
    default:
        log.Printf("同步错误: %v", err)
    }
}
```

### 获取同步状态

```go
// 获取定时同步状态，包括最后一次错误
status := ntp.GetPeriodicSyncStatus()
if status.LastError != nil {
    log.Printf("最后一次同步错误: %v", status.LastError)
}
```

## 高级用法

### 服务器管理

```go
// 添加新服务器
ntp.AddServer("time.cloudflare.com")

// 移除服务器
removed := ntp.RemoveServer("time.apple.com")
if removed {
    fmt.Println("服务器已移除")
}

// 获取当前服务器列表
servers := ntp.GetServers()
fmt.Printf("当前服务器: %v\n", servers)
```

### 自定义超时

```go
// 创建时设置超时
ntp, err := ntpsync.New(ntpsync.Options{
    Servers: []string{"pool.ntp.org"},
    Timeout: 3 * time.Second,
})

// 或者后续修改超时
ntp.SetTimeout(5 * time.Second)
```

### 异步同步

```go
// 异步执行同步（不等待结果）
ntp.SyncAsync()
```

## 最佳实践

### 服务器选择

- 使用地理位置接近的NTP服务器以减少网络延迟
- 使用公共NTP池（如pool.ntp.org）以获得更好的可用性
- 配置3-5个服务器以平衡可靠性和性能

### 同步间隔

- 对于一般应用，建议同步间隔为1小时
- 对于高精度要求，可以缩短到10-15分钟
- 避免过于频繁的同步（<1分钟），以减少网络负载

### 错误处理

- 始终检查同步错误
- 实现指数退避重试机制
- 监控同步状态和成功率

### 资源管理

- 在应用退出前调用`StopPeriodicSync()`以优雅关闭
- 避免创建多个NTP客户端实例，一个应用通常只需要一个实例
