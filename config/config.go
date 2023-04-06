package config

import (
	"fmt"
	"github.com/spf13/viper"
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
		Node              int    `mapstructure:"node"`                  // 雪花算法节点 ID
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
}

func InitConfig(configPath string) {
	fmt.Println("config init ...")
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

	fmt.Println("config init ok")
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
}
