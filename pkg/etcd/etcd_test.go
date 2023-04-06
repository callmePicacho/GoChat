package etcd

import (
	"GoChat/common"
	"GoChat/config"
	"GoChat/lib/etcd"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	config.InitConfig("../../app.yaml")

	// 创建一个新的 Discovery 实例
	srv, err := etcd.NewDiscovery()
	if err != nil {
		t.Fatalf("failed to create discovery: %v", err)
	}
	defer srv.Close()

	// 注册两个 k-v
	err = etcd.RegisterServer(common.EtcdServerList+"888", "888", 5)
	if err != nil {
		panic(err)
	}

	err = etcd.RegisterServer(common.EtcdServerList+"666", "666", 5)
	if err != nil {
		panic(err)
	}

	// 为一个存在的前缀启动 WatchService，并验证 GetServices 返回的服务列表是否正确
	go srv.WatchService(common.EtcdServerList)

	// 等待 watch
	time.Sleep(time.Second)

	services := srv.GetServices()
	if len(services) != 2 {
		t.Error("注册服务不足 2 个")
		for _, service := range services {
			t.Log(service)
		}
	}
}
