package cache

import (
	"GoChat/pkg/db"
	"GoChat/pkg/util"
	"context"
	"fmt"
	"time"
)

const (
	groupUserPrefix = "group_user_" // 群成员信息
	ttl2H           = 2 * 60 * 60   // 2h
)

func getGroupUserKey(groupId uint64) string {
	return fmt.Sprintf("%s%d", groupUserPrefix, groupId)
}

// SetGroupUser 设置群成员
func SetGroupUser(groupId uint64, userIds []uint64) error {
	key := getGroupUserKey(groupId)
	values := make([]string, 0, len(userIds))
	for _, userId := range userIds {
		values = append(values, util.Uint64ToStr(userId))
	}
	_, err := db.RDB.SAdd(context.Background(), key, values).Result()
	if err != nil {
		fmt.Println("[设置群成员信息] 错误,err:", err)
		return err
	}
	_, err = db.RDB.Expire(context.Background(), key, ttl2H*time.Second).Result()
	if err != nil {
		fmt.Println("[设置群成员信息] 过期时间设置错误,err:", err)
		return err
	}
	return nil
}

// GetGroupUser 获取群成员
func GetGroupUser(groupId uint64) ([]uint64, error) {
	key := getGroupUserKey(groupId)
	result, err := db.RDB.SMembers(context.Background(), key).Result()
	if err != nil {
		fmt.Println("[获取群成员信息] 错误，err:", err)
		return nil, err
	}
	userIds := make([]uint64, 0, len(result))
	for _, v := range result {
		userIds = append(userIds, util.StrToUint64(v))
	}
	return userIds, nil
}
