package db

import (
	"GoChat/pkg/logger"
	"context"
	"github.com/go-redis/redis/v8"
)

var (
	RDB *redis.Client
)

func InitRedis(addr, password string) {
	logger.Logger.Debug("Redis init ...")
	RDB = redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           0,
		Password:     password,
		PoolSize:     30,
		MinIdleConns: 30,
	})
	err := RDB.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}
	logger.Logger.Debug("Redis init ok")
}
