package rpc_server

import (
	"GoChat/config"
	"GoChat/pkg/protocol/pb"
	"GoChat/service/ws"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
)

type ConnectServer struct {
	pb.UnsafeConnectServer // 禁止向前兼容
}

func (*ConnectServer) DeliverMessage(ctx context.Context, req *pb.DeliverMessageReq) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	// 获取本地连接
	conn := ws.GetServer().GetConn(req.ReceiverId)
	if conn == nil || conn.GetUserId() != req.ReceiverId {
		fmt.Println("[DeliverMessage] 连接不存在 user_id:", req.ReceiverId)
		return resp, nil
	}

	// 消息发送
	conn.SendMsg(req.ReceiverId, req.Data)

	return resp, nil
}

func InitRPCServer() {
	rpcPort := config.GlobalConfig.App.RPCPort

	server := grpc.NewServer()
	pb.RegisterConnectServer(server, &ConnectServer{})

	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("rpc server 启动 ", rpcPort)

	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to rpc serve: %v", err)
	}
}
