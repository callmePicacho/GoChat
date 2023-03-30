package ws

import (
	"GoChat/common"
	"GoChat/lib/cache"
	"GoChat/model"
	"GoChat/pkg/protocol/pb"
	"GoChat/pkg/rpc"
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"
)

// GetOutputMsg 组装出下行消息
func GetOutputMsg(cmdType pb.CmdType, code int32, data []byte) ([]byte, error) {
	msg := &pb.Output{
		Type:    cmdType,
		Code:    code,
		CodeMsg: common.GetErrorMessage(common.OK, ""),
		Data:    data,
	}
	bytes, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("[心跳] proto.Marshal err:", err)
		return nil, err
	}
	return bytes, nil
}

// SendToUser 发送消息到好友
func SendToUser(msg *pb.Message, bytes []byte) error {
	// 消息存储
	err := model.CreateMessage(&model.Message{
		UserID:      msg.SenderId,
		SessionType: int8(msg.SessionType),
		ReceiverId:  msg.ReceiverId,
		MessageType: int8(msg.MessageType),
		Content:     string(msg.Content),
	})
	if err != nil {
		fmt.Println("[消息处理] 存储失败，err:", err)
		return err
	}

	// 消息转发
	// 是否在线 ---否---> 离线消息存储 TODO
	//    |
	//    是
	//    ↓
	//  是否在本地 --否--> RPC 调用  -->  是否在本地   --否--> TODO
	//    |                               |
	//    是                              是
	//    ↓                               ↓
	//  消息发送                          消息发送

	// 查询是否在线
	rpcAddr, err := cache.GetUserOnline(msg.ReceiverId)
	if err != nil {
		fmt.Println("[消息处理]，用户不在线，msg:", msg)
		return err
	}
	// 不在线
	if rpcAddr == "" {
		// TODO 离线消息存储
		return nil
	}

	fmt.Println("[消息处理] 用户在线，rpcAddr:", rpcAddr)

	// 查询是否在本地
	conn := ConnManager.GetConn(msg.ReceiverId)
	if conn != nil {
		// 发送本地消息
		conn.SendMsg(msg.ReceiverId, bytes)
		fmt.Println("[消息处理]， 发送本地消息给用户, ", msg.ReceiverId)
		return nil
	}

	// rpc 调用
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = rpc.GetServerClient(rpcAddr).DeliverMessage(ctx, &pb.DeliverMessageReq{
		ReceiverId: msg.ReceiverId,
		Data:       bytes,
	})

	if err != nil {
		fmt.Println("[消息处理] DeliverMessage err, err:", err)
		return err
	}
	return nil
}

// SendToGroup 发送消息到群
func SendToGroup() {

}
