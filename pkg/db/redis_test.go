package db

import (
	"GoChat/config"
	"testing"
)

func TestRedis(t *testing.T) {
	config.InitConfig("../../app.yaml")
	InitRedis(config.GlobalConfig.Redis.Addr, config.GlobalConfig.Redis.Password)

}
