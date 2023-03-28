package router

import (
	"GoChat/pkg/middlewares"
	"GoChat/service"
	"github.com/gin-gonic/gin"
)

// HTTPRouter http 路由
func HTTPRouter() *gin.Engine {
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

	return r
}
