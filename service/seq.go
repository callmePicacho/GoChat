package service

import (
	"GoChat/lib/cache"
)

func GetUserNextSeq(userId uint64) (uint64, error) {
	return cache.GetNextSeqId(cache.SeqObjectTypeUser, userId)
}

func GetUserNextSeqBatch(userIds []uint64) ([]uint64, error) {
	return cache.GetNextSeqIds(cache.SeqObjectTypeUser, userIds)
}
