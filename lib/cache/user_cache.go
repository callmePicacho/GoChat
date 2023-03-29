package cache

import (
	"GoChat/pkg/db"
	"context"
	"fmt"
	"time"
)

const (
	userOnlinePrefix = "user_online_"   // 用户在线状态设置
	ttl7D            = 7 * 24 * 60 * 60 // s  7天
)

func GetUserKey(userId uint64) string {
	return fmt.Sprintf("%s%d", userOnlinePrefix, userId)
}

// SetUserOnline 设置用户在线
func SetUserOnline(userId uint64, addr string) error {
	key := GetUserKey(userId)
	_, err := db.RDB.Set(context.Background(), key, addr, ttl7D*time.Second).Result()
	if err != nil {
		fmt.Println("[设置用户在线] 错误, err:", err)
	}
	return err
}

// DelUserOnline 删除用户在线信息（存在即在线）
func DelUserOnline(userId uint64) error {
	key := GetUserKey(userId)
	_, err := db.RDB.Del(context.Background(), key).Result()
	if err != nil {
		fmt.Println("[删除用户在线] 错误, err:", err)
	}
	return err
}
