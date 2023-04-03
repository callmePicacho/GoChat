package service

import "GoChat/model"

func GetUserNextSeq(userId uint64) (uint64, error) {
	return model.Incr(model.SeqObjectTypeUser, userId)
}
