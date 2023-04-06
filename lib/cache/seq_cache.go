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
func GetNextSeqIds(objectType int8, objectIds []uint64) ([]uint64, error) {
	script := `
       local results = {}
       for i, key in ipairs(KEYS) do
           results[i] = redis.call('INCR', key)
       end
       return results
   `
	keys := make([]string, len(objectIds))
	for i, objectId := range objectIds {
		keys[i] = getSeqKey(objectType, objectId)
	}
	res, err := redis.NewScript(script).Run(context.Background(), db.RDB, keys).Result()
	if err != nil {
		fmt.Println("[获取seq] 失败，err:", err)
		return nil, err
	}
	results := make([]uint64, len(objectIds))
	for i, v := range res.([]interface{}) {
		results[i] = uint64(v.(int64))
	}
	return results, nil
}
