package service

import (
	"GoChat/model"
	"GoChat/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AddFriend 添加好友
func AddFriend(c *gin.Context) {
	// 参数验证
	friendIdStr := c.PostForm("friend_id")
	friendId := util.StrToUint64(friendIdStr)
	if friendId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	// 获取自己的信息
	uc := c.MustGet("user_claims").(*util.UserClaims)
	if uc.UserId == friendId {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "不能添加自己为好友",
		})
		return
	}
	// 查询用户是否存在
	ub, err := model.GetUserById(friendId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "好友不存在",
		})
		return
	}
	// 查询是否已建立好友关系
	isFriend, err := model.IsFriend(uc.UserId, ub.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	if isFriend {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "请勿重复添加",
		})
		return
	}

	// 建立好友关系
	friend := &model.Friend{
		UserID:   uc.UserId,
		FriendID: ub.ID,
	}
	err = model.CreateFriend(friend)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "好友添加成功",
	})
}
