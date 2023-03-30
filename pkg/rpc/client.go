package rpc

import (
	"GoChat/pkg/protocol/pb"
	"fmt"
	"google.golang.org/grpc"
)

var (
	ConnServerClient pb.ConnectClient
)

// GetServerClient 获取 grpc 连接
func GetServerClient(addr string) pb.ConnectClient {
	client, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("grpc client Dial err, err:", err)
		panic(err)
	}
	ConnServerClient = pb.NewConnectClient(client)
	return ConnServerClient
}
