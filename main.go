package main

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"GoChat/router"
)

func main() {
	// 初始化
	config.InitConfig()
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	r := router.HTTPRouter()

	r.Run(":" + config.GlobalConfig.App.HTTPServerPort)
}
