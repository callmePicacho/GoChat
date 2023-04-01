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
		Content:     msg.Content,
	})
	if err != nil {
		fmt.Println("[私聊消息处理] 存储失败，err:", err)
		return err
	}

	// 进行推送
	return Send(msg.ReceiverId, bytes)
}

// SendToGroup 发送消息到群
func SendToGroup(msg *pb.Message, bytes []byte) error {
	// 获取群成员信息
	userIds, err := model.GetGroupUserIdsByGroupId(msg.ReceiverId)
	if err != nil {
		fmt.Println("[群聊消息处理] 查询失败，err:", err, msg)
		return err
	}

	// 检查当前用户是否属于该群
	isMember := false
	for _, userId := range userIds {
		if msg.SenderId == userId {
			isMember = true
			break
		}
	}

	if !isMember {
		fmt.Println("[群聊消息处理] 用户不属于该群组，msg:", msg)
		return nil
	}

	// 存储数据
	err = model.CreateGroupMsg(&model.GroupMsg{
		UserID:      msg.SenderId,
		GroupID:     msg.ReceiverId,
		MessageType: int8(msg.MessageType),
		SendTime:    time.Now(), // TODO
		Content:     msg.Content,
	})
	if err != nil {
		fmt.Println("[群聊消息处理] 存储失败，err:", err)
		return err
	}
	// 进行推送
	for _, userId := range userIds {
		if userId == msg.SenderId {
			continue
		}
		err = Send(userId, bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// Send 消息转发
// 是否在线 ---否---> 离线消息存储 TODO
//    |
//    是
//    ↓
//  是否在本地 --否--> RPC 调用  -->  是否在本地   --否--> TODO
//    |                               |
//    是                              是
//    ↓                               ↓
//  消息发送                          消息发送

func Send(receiverId uint64, bytes []byte) error {
	// 查询是否在线
	rpcAddr, err := cache.GetUserOnline(receiverId)
	if err != nil {
		return err
	}

	// 不在线
	if rpcAddr == "" {
		fmt.Println("[消息处理]，用户不在线，receiverId:", receiverId)
		// TODO 离线消息存储
		return nil
	}

	fmt.Println("[消息处理] 用户在线，rpcAddr:", rpcAddr)

	// 查询是否在本地
	conn := ConnManager.GetConn(receiverId)
	if conn != nil {
		// 发送本地消息
		conn.SendMsg(receiverId, bytes)
		fmt.Println("[消息处理]， 发送本地消息给用户, ", receiverId)
		return nil
	}

	// rpc 调用
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = rpc.GetServerClient(rpcAddr).DeliverMessage(ctx, &pb.DeliverMessageReq{
		ReceiverId: receiverId,
		Data:       bytes,
	})

	if err != nil {
		fmt.Println("[消息处理] DeliverMessage err, err:", err)
		return err
	}

	return nil
}
