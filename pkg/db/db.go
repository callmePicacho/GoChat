package db

import (
	"GoChat/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"moul.io/zapgorm2"
	"time"
)

var (
	DB *gorm.DB
)

func InitMySQL(dataSource string) {
	logger.Logger.Info("mysql init...")
	var err error
	newLogger := zapgorm2.New(logger.Logger)
	newLogger.SetAsDefault()

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
	logger.Logger.Info("mysql init ok")
}
