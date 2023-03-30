package router

import (
	"GoChat/config"
	"GoChat/pkg/middlewares"
	"GoChat/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// HTTPRouter http 路由
func HTTPRouter() {
	r := gin.Default()

	// 用户注册
	r.POST("/register", service.Register)

	// 用户登录
	r.POST("/login", service.Login)

	auth := r.Group("", middlewares.AuthCheck())
	{
		// 添加好友
		auth.POST("/friend/add", service.AddFriend)

		// 创建群聊
		auth.POST("/group/create", service.CreateGroup)

		// 获取群成员列表
		auth.GET("/group_user/list", service.GroupUserList)
	}

	httpAddr := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.HTTPServerPort)
	if err := r.Run(httpAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
