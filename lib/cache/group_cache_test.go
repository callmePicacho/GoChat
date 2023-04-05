package cache

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"testing"
)

func TestGroupUser(t *testing.T) {
	config.InitConfig("../../app.yaml")
	db.InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)
	var groupId uint64 = 77777
	_, err := GetGroupUser(groupId)
	if err != nil {
		panic(err)
	}
}
