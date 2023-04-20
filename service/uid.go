package service

import (
	"GoChat/pkg/util"
)

func GetUserNextId(userId uint64) (uint64, error) {
	return util.UidGen.GetNextId(util.Uint64ToStr(userId))
}

// GetUserNextIdBatch 批量获取 seq
func GetUserNextIdBatch(userIds []uint64) ([]uint64, error) {
	businessIds := make([]string, len(userIds))
	for i, userId := range userIds {
		businessIds[i] = util.Uint64ToStr(userId)
	}
	return util.UidGen.GetNextIds(businessIds)
}
