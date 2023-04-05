package service

import (
	"GoChat/lib/cache"
	"github.com/go-redis/redis/v8"
)

func GetUserNextSeq(userId uint64) (uint64, error) {
	return cache.GetNextSeqId(cache.SeqObjectTypeUser, userId)
}

// GetUserNextSeqWithPipeline 使用 pipeline 封装
func GetUserNextSeqWithPipeline(pipeliner redis.Pipeliner, userIds []uint64) ([]uint64, error) {
	return cache.GetNextSeqIds(pipeliner, cache.SeqObjectTypeUser, userIds)
}
