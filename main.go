package main

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"GoChat/router"
	"GoChat/service/rpc_server"
)

func main() {
	// 初始化
	config.InitConfig()
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	// 启动 http 服务
	go router.HTTPRouter()

	// 启动 rpc 服务
	go rpc_server.InitRPCServer()

	// 启动 websocket 服务
	router.WSRouter()
}
