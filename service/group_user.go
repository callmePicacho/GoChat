package service

import (
	"GoChat/model"
	"GoChat/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GroupUserList 获取群成员列表
func GroupUserList(c *gin.Context) {
	// 参数校验
	groupIdStr := c.Query("group_id")
	groupId := util.StrToUint64(groupIdStr)
	if groupId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	// 获取用户信息
	uc := c.MustGet("user_claims").(*util.UserClaims)

	// 验证用户是否属于该群
	isBelong, err := model.IsBelongToGroup(uc.UserId, groupId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	if !isBelong {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户不属于该群",
		})
		return
	}

	// 获取群成员id列表
	ids, err := model.GetGroupUserIdsByGroupId(groupId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, util.Uint64ToStr(id))
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "请求成功",
		"data": gin.H{
			"ids": idsStr,
		},
	})
}
