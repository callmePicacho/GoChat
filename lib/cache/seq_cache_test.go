package cache

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"testing"
)

func TestGetNextSeqIds(t *testing.T) {
	config.InitConfig("../../app.yaml")
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

	userIds := []uint64{1, 2, 3, 4, 5}
	ids, err := GetNextSeqIds(SeqObjectTypeUser, userIds)
	if err != nil {
		panic(err)
	}
	t.Log(ids)
}
