/*
Package ntpsync 提供NTP时间同步功能。

本包使用纯Go语言实现NTP（网络时间协议）客户端功能，
不依赖任何第三方库。它支持多个NTP服务器，具有自动故障转移、
定时同步和服务器状态监控功能。

基本用法：

    // 创建新的NTP客户端
    ntp, err := ntpsync.New(ntpsync.Options{
        Servers: []string{"pool.ntp.org"},
    })
    if err != nil {
        // 处理错误
    }

    // 执行同步
    err = ntp.Sync()
    if err != nil {
        // 处理错误
    }

    // 获取经NTP调整后的当前时间
    ntpTime := ntp.Now()

多服务器支持：

    // 创建支持多服务器的客户端
    ntp, err := ntpsync.New(ntpsync.Options{
        Servers: []string{
            "pool.ntp.org",
            "time.google.com",
            "time.windows.com",
        },
        EnableMultiServer: true,
    })

    // 获取所有服务器的状态
    statuses, err := ntp.GetMultiServerStatus()

定时同步：

    // 创建具有自动同步功能的客户端
    ntp, err := ntpsync.New(ntpsync.Options{
        Servers: []string{"pool.ntp.org"},
        SyncInterval: 1 * time.Hour,
        AutoSync: true,
    })

    // 或手动启动定时同步
    err = ntp.StartPeriodicSync()

    // 停止定时同步
    ntp.StopPeriodicSync()

服务器管理：

    // 添加新服务器
    ntp.AddServer("time.cloudflare.com")

    // 移除服务器
    removed := ntp.RemoveServer("time.apple.com")

    // 获取当前服务器列表
    servers := ntp.GetServers()

更多信息，请参阅README.md文件。
*/
package ntpsync
