package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var (
	DB *gorm.DB
)

func InitMySQL(dataSource string) {
	fmt.Println("MySQL init...")
	var err error
	//logFile, err := os.Create("./log/err.log")
	//if err != nil {
	//	panic(err)
	//}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,  // 慢SQL阈值
			Colorful:      true,         // 颜色
			LogLevel:      logger.Error, // 级别
		},
	)

	DB, err = gorm.Open(mysql.Open(dataSource),
		&gorm.Config{
			Logger: newLogger,
		})
	if err != nil {
		panic(err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		panic(err)
	}
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量
	sqlDB.SetMaxIdleConns(20)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(30)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)
	fmt.Println("MySQL init ok")
}
