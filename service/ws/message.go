package ws

import (
	"GoChat/common"
	"GoChat/lib/cache"
	"GoChat/model"
	"GoChat/pkg/protocol/pb"
	"GoChat/pkg/rpc"
	"GoChat/pkg/util"
	"GoChat/service"
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
func SendToUser(msg *pb.Message, userId uint64) error {
	// 获取接受者 seqId
	seq, err := service.GetUserNextSeq(userId)
	if err != nil {
		fmt.Println("[消息处理] 获取 seq 失败,err:", err)
		return err
	}
	msg.Seq = seq

	// 消息存储
	err = model.CreateMessage(&model.Message{
		UserID:      userId,
		SenderID:    msg.SenderId,
		SessionType: int8(msg.SessionType),
		ReceiverId:  msg.ReceiverId,
		MessageType: int8(msg.MessageType),
		Content:     msg.Content,
		Seq:         seq,
	})
	if err != nil {
		fmt.Println("[消息处理] 存储失败，err:", err)
		return err
	}

	// 如果发给自己的，只落库不进行发送
	if userId == msg.SenderId {
		return nil
	}

	// 组装消息
	pushMsg := &pb.PushMsg{Msg: msg}
	pushMsgBytes, err := proto.Marshal(pushMsg)
	if err != nil {
		fmt.Println("[消息处理] pushMsg Marshal error,err:", err)
		return err
	}
	bytes, err := GetOutputMsg(pb.CmdType_CT_Message, int32(common.OK), pushMsgBytes)
	if err != nil {
		fmt.Println("[消息处理] GetOutputMsg Marshal error,err:", err)
		return err
	}

	// 进行推送
	return Send(userId, bytes)
}

// SendToGroup 发送消息到群
func SendToGroup(msg *pb.Message) error {
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

	go func() {
		defer util.RecoverPanic()
		// 进行推送
		for _, userId := range userIds {
			if userId == msg.SenderId {
				continue
			}
			err := SendToUser(msg, userId)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

// Send 消息转发
// 是否在线 ---否---> 不进行推送
//    |
//    是
//    ↓
//  是否在本地 --否--> RPC 调用  -->  是否在本地
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
