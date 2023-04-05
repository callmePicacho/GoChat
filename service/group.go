package service

import (
	"GoChat/lib/cache"
	"GoChat/model"
	"GoChat/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateGroup 创建群聊
func CreateGroup(c *gin.Context) {
	// 参数校验
	name := c.PostForm("name")
	idsStr := c.PostFormArray("ids") // 群成员 id，不包括群创建者
	if name == "" || len(idsStr) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	ids := make([]uint64, 0, len(idsStr)+1)
	for i := range idsStr {
		ids = append(ids, util.StrToUint64(idsStr[i]))
	}
	// 获取用户信息
	uc := c.MustGet("user_claims").(*util.UserClaims)
	ids = append(ids, uc.UserId)

	// 获取 ids 用户信息
	ids, err := model.GetUserIdByIds(ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 创建群组
	group := &model.Group{
		Name:    name,
		OwnerID: uc.UserId,
	}
	err = model.CreateGroup(group, ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	// 将群成员信息更新到 Redis
	err = cache.SetGroupUser(group.ID, ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "群组创建成功",
		"data": gin.H{
			"id": util.Uint64ToStr(group.ID),
		},
	})
}
