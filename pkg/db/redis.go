package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var (
	RDB *redis.Client
)

func InitRedis(addr, password string) {
	fmt.Println("Redis init...")
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: password,
	})
	err := RDB.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis init ok")
}
