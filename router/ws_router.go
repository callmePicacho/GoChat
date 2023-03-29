package router

import (
	"GoChat/service/ws"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
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
	server := ws.NewServer()

	// 开启worker工作池
	server.StartWorkerPool()

	// 开启心跳超时检测
	checker := ws.NewHeartbeatChecker(10*time.Second, server)
	go checker.Start()

	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		// 升级协议  http -> websocket
		WsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("websocket conn err :", err)
			return
		}

		// 初始化连接
		conn := ws.NewConnection(server, WsConn)

		// 开启读写线程
		go conn.Start()
	})

	r.Run(server.Addr())
}
