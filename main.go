package main

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"GoChat/router"
	"GoChat/service/rpc_server"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 初始化
	config.InitConfig()
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	// 启动 http 服务
	go func() {
		r := router.HTTPRouter()
		httpAddr := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.HTTPServerPort)
		if err := r.Run(httpAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 启动 rpc 服务
	go rpc_server.InitRPCServer()

	// 启动 websocket 服务
	router.WSRouter()
}
