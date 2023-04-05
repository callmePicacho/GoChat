package cache

import (
	"GoChat/pkg/db"
	"context"
	"fmt"
)

const (
	seqPrefix = "object_seq_" // 群成员信息

	SeqObjectTypeUser = 1 // 用户
)

// 消息同步序列号
func getSeqKey(objectType int8, objectId uint64) string {
	return fmt.Sprintf("%s%d_%d", seqPrefix, objectType, objectId)
}

// GetNextSeqId 获取用户的下一个 seq，消息同步序列号
func GetNextSeqId(objectType int8, objectId uint64) (uint64, error) {
	key := getSeqKey(objectType, objectId)
	result, err := db.RDB.Incr(context.Background(), key).Uint64()
	if err != nil {
		fmt.Println("[获取seq] 失败，err:", err)
		return 0, err
	}
	return result, nil
}
