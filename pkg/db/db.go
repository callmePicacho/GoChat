package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func InitMySQL(dataSource string) {
	fmt.Println("MySQL init...")
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢SQL阈值
			Colorful:      true,        // 颜色
			LogLevel:      logger.Info, // 级别
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dataSource),
		&gorm.Config{Logger: newLogger})
	if err != nil {
		panic(err)
	}
	fmt.Println("MySQL init ok")
}

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
