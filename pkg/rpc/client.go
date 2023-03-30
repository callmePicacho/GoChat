package rpc

import (
	"GoChat/config"
	"GoChat/pkg/protocol/pb"
	"fmt"
	"google.golang.org/grpc"
	"sync"
)

var (
	ConnServerClient pb.ConnectClient
	once             sync.Once
)

// GetServerClient 获取 grpc 连接
func GetServerClient() pb.ConnectClient {
	once.Do(func() {
		addr := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.RPCPort)
		client, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			fmt.Println("grpc client Dial err, err:", err)
			panic(err)
		}
		ConnServerClient = pb.NewConnectClient(client)
	})
	return ConnServerClient
}
