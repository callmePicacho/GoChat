package etcd

import (
	"GoChat/common"
	"GoChat/config"
	"GoChat/lib/etcd"
	"GoChat/pkg/logger"
	"fmt"
	"go.uber.org/zap"
	"time"
)

var (
	DiscoverySer *etcd.Discovery
)

// InitETCD 初始化服务注册发现
// 1. 初始化服务注册，将自己当前启动的 RPC 端口注册到 etcd
// 2. 初始化服务发现，启动 watcher 监听所有 RPC 端口，以便有需要时能直接获取当前注册在 ETCD 的服务
func InitETCD() {
	hostPort := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.RPCPort)
	logger.Logger.Info("注册服务", zap.String("hostport", hostPort))

	// 注册服务并设置 k-v 租约
	err := etcd.RegisterServer(common.EtcdServerList+hostPort, hostPort, 5)
	if err != nil {
		return
	}

	time.Sleep(100 * time.Millisecond)

	DiscoverySer, err = etcd.NewDiscovery()
	if err != nil {
		return
	}

	// 阻塞监听
	DiscoverySer.WatchService(common.EtcdServerList)
}
