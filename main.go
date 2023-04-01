package main

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"GoChat/router"
	"GoChat/service/rpc_server"
)

func main() {
	// 初始化
	config.InitConfig("./app.yaml")
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	// 初始化服务注册发现
	//go etcd.InitETCD()

	// 启动 http 服务
	go router.HTTPRouter()

	// 启动 rpc 服务
	go rpc_server.InitRPCServer()

	// 启动 websocket 服务
	router.WSRouter()
}
