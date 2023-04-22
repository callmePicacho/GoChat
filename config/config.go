package config

import (
	"GoChat/pkg/logger"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var GlobalConfig *Configuration

type Configuration struct {
	// JWT 配置
	JWT struct {
		SignKey    string `mapstructure:"sign_key"`    // JWT 签名密钥
		ExpireTime int    `mapstructure:"expire_time"` // JWT 过期时间（小时）
	} `mapstructure:"jwt"`

	// MySQL 配置
	MySQL struct {
		DNS string `mapstructure:"dns"` // 数据库连接字符串
	} `mapstructure:"mysql"`

	// Redis 配置
	Redis struct {
		Addr     string `mapstructure:"addr"`     // Redis 地址
		Password string `mapstructure:"password"` // Redis 认证密码
	} `mapstructure:"redis"`

	// 应用程序配置
	App struct {
		Salt              string `mapstructure:"salt"`                  // 密码加盐
		IP                string `mapstructure:"ip"`                    // 应用程序 IP 地址
		HTTPServerPort    string `mapstructure:"http_server_port"`      // HTTP 服务器端口
		WebsocketPort     string `mapstructure:"websocket_server_port"` // WebSocket 服务器端口
		RPCPort           string `mapstructure:"rpc-port"`              // RPC 服务器端口
		WorkerPoolSize    uint32 `mapstructure:"worker_pool_size"`      // 业务 worker 队列数量
		MaxWorkerTask     int    `mapstructure:"max_worker_task"`       // 业务 worker 对应负责的任务队列最大任务存储数量
		HeartbeatTimeout  int    `mapstructure:"heartbeattime"`         // 心跳超时时间（秒）
		HeartbeatInterval int    `mapstructure:"heartbeatInterval"`     // 超时连接检测间隔（秒）
	} `mapstructure:"app"`

	// ETCD相关配置
	ETCD struct {
		Endpoints []string `mapstructure:"endpoints"` // etcd endpoints 列表
		Timeout   int      `mapstructure:"timeout"`   // 超时时间（秒）
	} `mapstructure:"etcd"`

	RabbitMQ struct {
		URL string `mapstructure:"url"` // rabbitmq url
	}

	Log struct {
		Target   string `mapstructure:"target"` // 日志输出路径：可选值 console/file
		Level    string `mapstructure:"level"`  // 日志输出级别
		LevelNum zapcore.Level
	}
}

func (c Configuration) String() string {
	return fmt.Sprintf(
		"JWT:\n  SignKey: %s\n  ExpireTime: %d\nMySQL:\n  DNS: %s\nRedis:\n  Addr: %s\n  Password: %s\nApp:\n  Salt: %s\n  IP: %s\n  HTTPServerPort: %s\n  WebsocketPort: %s\n  RPCPort: %s\n  WorkerPoolSize: %d\n  MaxWorkerTask: %d\n  HeartbeatTimeout: %d\n  HeartbeatInterval: %d\nETCD:\n  Endpoints: %v\n  Timeout: %d\nRabbitMQ:\n  URL: %s\nLog:\n  Target: %s\n  Level: %s\n",
		c.JWT.SignKey,
		c.JWT.ExpireTime,
		c.MySQL.DNS,
		c.Redis.Addr,
		c.Redis.Password,
		c.App.Salt,
		c.App.IP,
		c.App.HTTPServerPort,
		c.App.WebsocketPort,
		c.App.RPCPort,
		c.App.WorkerPoolSize,
		c.App.MaxWorkerTask,
		c.App.HeartbeatTimeout,
		c.App.HeartbeatInterval,
		c.ETCD.Endpoints,
		c.ETCD.Timeout,
		c.RabbitMQ.URL,
		c.Log.Target,
		c.Log.Level,
	)
}

func InitConfig(configPath string) {
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	GlobalConfig = new(Configuration)
	err = viper.Unmarshal(GlobalConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
	reload()

	// 初始化 log
	logger.InitLogger(GlobalConfig.Log.Target, GlobalConfig.Log.LevelNum)
	logger.Logger.Debug("config  init ok", zap.String("GlobalConfig", GlobalConfig.String()))
}

func reload() {
	// 最小为 10
	if GlobalConfig.App.WorkerPoolSize < 10 {
		GlobalConfig.App.WorkerPoolSize = 10
	}
	// 最小为 1024
	if GlobalConfig.App.MaxWorkerTask < 1000 {
		GlobalConfig.App.MaxWorkerTask = 1024
	}
	// 默认为控制台
	if GlobalConfig.Log.Target == "file" {
		GlobalConfig.Log.Target = logger.File
	} else {
		GlobalConfig.Log.Target = logger.Console
	}
	// 如果解析失败默认为 Error 级别
	var err error
	GlobalConfig.Log.LevelNum, err = zapcore.ParseLevel(GlobalConfig.Log.Level)
	if err != nil {
		GlobalConfig.Log.LevelNum = zapcore.ErrorLevel
	}
	fmt.Println("日志级别为：", GlobalConfig.Log.LevelNum)
	fmt.Println("日志输出到：", GlobalConfig.Log.Target)
}
