package service

import (
	"GoChat/model"
	"GoChat/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Register 注册
func Register(c *gin.Context) {
	// 获取参数并验证
	phoneNumber := c.PostForm("phone_number")
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	if phoneNumber == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	// 查询手机号是否已存在
	cnt, err := model.GetUserCountByPhone(nickname)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "账号已被注册",
		})
		return
	}
	// 插入用户信息
	ub := &model.User{
		PhoneNumber: phoneNumber,
		Nickname:    nickname,
		Password:    util.GetMD5(password),
	}
	err = model.CreateUser(ub)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误" + err.Error(),
		})
		return
	}

	// 生成 token
	token, err := util.GenerateToken(ub.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 发放 token
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
		},
	})
}

// Login 登录
func Login(c *gin.Context) {
	// 验证参数
	phoneNumber := c.PostForm("phone_number")
	password := c.PostForm("password")
	if phoneNumber == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	// 验证账号名和密码是否正确
	ub, err := model.GetUserByPhoneAndPassword(phoneNumber, util.GetMD5(password))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "手机号或密码错误",
		})
		return
	}
	// 生成 token
	token, err := util.GenerateToken(ub.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 发放 token
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token":   token,
			"user_id": util.Uint64ToStr(ub.ID),
		},
	})
}
