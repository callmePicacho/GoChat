package router

import (
	"GoChat/config"
	"GoChat/service/ws"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSRouter websocket 路由
func WSRouter() {
	server := ws.GetServer()

	// 开启worker工作池
	server.StartWorkerPool()

	// 开启心跳超时检测
	checker := ws.NewHeartbeatChecker(time.Second*time.Duration(config.GlobalConfig.App.HeartbeatInterval), server)
	go checker.Start()

	r := gin.Default()

	var connID uint64

	r.GET("/ws", func(c *gin.Context) {
		// 升级协议  http -> websocket
		WsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("websocket conn err :", err)
			return
		}

		// 初始化连接
		conn := ws.NewConnection(server, WsConn, connID)
		connID++

		// 开启读写线程
		go conn.Start()
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.WebsocketPort),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 关闭服务
	server.Stop()

	// 5s 超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown: ", err)
	}

	log.Println("Server exiting")
}
