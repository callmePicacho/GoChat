package cache

import (
	"GoChat/pkg/db"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
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

// GetNextSeqIds 获取多个对象的下一个 seq，消息同步序列号
func GetNextSeqIds(pipeliner redis.Pipeliner, objectType int8, objectIds []uint64) ([]uint64, error) {
	results := make([]uint64, len(objectIds))
	for _, objectId := range objectIds {
		key := getSeqKey(objectType, objectId)
		pipeliner.Incr(context.Background(), key)
	}
	cmds, err := pipeliner.Exec(context.Background())
	if err != nil {
		fmt.Println("[获取seq] 失败，err:", err)
		return nil, err
	}
	for i, cmd := range cmds {
		result, err := cmd.(*redis.IntCmd).Uint64()
		if err != nil {
			fmt.Println("[获取seq] 失败，err:", err)
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}
