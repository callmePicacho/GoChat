package main

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"GoChat/router"
	"fmt"
)

func main() {
	// 初始化
	config.InitConfig()
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	// http 服务
	r := router.HTTPRouter()

	// websocket 服务
	go router.WSRouter()

	r.Run(fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.HTTPServerPort))
}
