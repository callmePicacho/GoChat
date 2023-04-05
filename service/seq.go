package service

import (
	"GoChat/lib/cache"
)

func GetUserNextSeq(userId uint64) (uint64, error) {
	return cache.GetNextSeqId(cache.SeqObjectTypeUser, userId)
}
