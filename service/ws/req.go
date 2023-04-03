package ws

import (
	"GoChat/common"
	"GoChat/config"
	"GoChat/lib/cache"
	"GoChat/model"
	"GoChat/pkg/protocol/pb"
	"GoChat/pkg/util"
	"fmt"
	"google.golang.org/protobuf/proto"
)

// Handler 路由函数
type Handler func()

// Req 请求
type Req struct {
	conn *Conn   // 连接
	data []byte  // 客户端发送的请求数据
	f    Handler // 该请求需要执行的路由函数
}

func (r *Req) Login() {

	// 检查用户是否已登录 只能防止同一个连接多次调用 Login
	// TODO 要防止多个连接使用相同 token 进行 Login，还需要验证 Redis 中是否存在用户数据并做相应处理
	if r.conn.GetUserId() != 0 {
		fmt.Println("[用户登录] 用户已登录")
		return
	}

	// 消息解析 proto string -> struct
	loginMsg := new(pb.LoginMsg)
	err := proto.Unmarshal(r.data, loginMsg)
	if err != nil {
		fmt.Println("[用户登录] unmarshal error,err:", err)
		return
	}
	// 登录校验
	userClaims, err := util.AnalyseToken(string(loginMsg.Token))
	if err != nil {
		fmt.Println("[用户登录] AnalyseToken err:", err)
		return
	}

	// 设置 user_id
	r.conn.SetUserId(userClaims.UserId)

	// Redis 存储用户数据 k: userId,  v: grpc地址，方便用户能直接通过这个地址进行 rpc 方法调用
	grpcServerAddr := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.RPCPort)
	err = cache.SetUserOnline(userClaims.UserId, grpcServerAddr)
	if err != nil {
		fmt.Println("[用户登录] 系统错误")
		return
	}

	// 回复ACK
	ackMsg := &pb.ACKMsg{
		Type:     pb.ACKType_AT_Login,
		ClientId: 0, // TODO 回复 clientID 给客户端使用
	}
	ackMsgBytes, err := proto.Marshal(ackMsg)
	if err != nil {
		fmt.Println("[用户登录] 系统错误")
		return
	}

	bytes, err := GetOutputMsg(pb.CmdType_CT_ACK, int32(common.OK), ackMsgBytes)
	if err != nil {
		fmt.Println("[用户登录] proto.Marshal err:", err)
		return
	}

	// 回复发送 Login 请求的客户端
	r.conn.SendMsg(userClaims.UserId, bytes)

	// 加入到 connMap 中
	r.conn.server.AddConn(userClaims.UserId, r.conn)
}

func (r *Req) HeartBeat() {
	// TODO 更新当前用户状态，不做回复
}

// MessageHandler 消息处理，处理客户端发送给服务端的消息
// A 发送消息给服务端，服务端收到消息处理后发给 B
// 包括：单聊、群聊
func (r *Req) MessageHandler() {
	// 消息解析 proto string -> struct
	msg := new(pb.UpMsg)
	err := proto.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[消息处理] unmarshal error,err:", err)
		return
	}

	// 实现消息可靠性
	if !r.conn.CompareAndIncrClientID(msg.ClientId) {
		fmt.Println("不是想要收到的 clientID，不进行处理, msg:", msg)
		return
	}

	if msg.Msg.SenderId != r.conn.GetUserId() {
		fmt.Println("[消息处理] 发送者有误")
		return
	}

	// 单聊不能发给自己
	if msg.Msg.SessionType == pb.SessionType_ST_Single && msg.Msg.ReceiverId == r.conn.GetUserId() {
		fmt.Println("[消息处理] 接收者有误")
		return
	}

	// 给自己发一份，消息落库但是不推送 TODO 需要回复 seq
	err = SendToUser(msg.Msg, msg.Msg.SenderId)
	if err != nil {
		fmt.Println("[消息处理] send to 自己出错, err:", err)
		return
	}

	// 单聊、群聊
	switch msg.Msg.SessionType {
	case pb.SessionType_ST_Single:
		err = SendToUser(msg.Msg, msg.Msg.ReceiverId)
	case pb.SessionType_ST_Group:
		err = SendToGroup(msg.Msg)
	default:
		fmt.Println("[消息处理] 会话类型错误")
		return
	}
	if err != nil {
		fmt.Println("[消息处理] 系统错误")
		return
	}

	// 回复发送上行消息的客户端 ACK
	ackMsg := &pb.ACKMsg{
		Type:     pb.ACKType_AT_Up,
		ClientId: msg.ClientId, // 回复客户端，当前已 ACK 的消息
	}
	ackMsgBytes, err := proto.Marshal(ackMsg)
	if err != nil {
		fmt.Println("[消息处理] 系统错误")
		return
	}

	ackBytes, err := GetOutputMsg(pb.CmdType_CT_ACK, int32(common.OK), ackMsgBytes)
	if err != nil {
		fmt.Println("[消息处理] proto.Marshal err:", err)
		return
	}
	// 回复发送 Message 请求的客户端 A
	r.conn.SendMsg(msg.Msg.SenderId, ackBytes)
}

// Sync  消息同步，拉取离线消息
func (r *Req) Sync() {
	msg := new(pb.SyncInputMsg)
	err := proto.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[离线消息] unmarshal error,err:", err)
		return
	}

	// 根据 seq 查询，得到比 seq 大的用户消息
	messages, hasMore, err := model.ListByUserIdAndSeq(r.conn.GetUserId(), msg.Seq, model.MessageLimit)
	if err != nil {
		fmt.Println("[离线消息] model.ListByUserIdAndSeq error, err:", err)
		return
	}
	pbMessage := model.MessagesToPB(messages)

	syncOut := &pb.SyncOutputMsg{
		Messages: pbMessage,
		HasMore:  hasMore,
	}
	syncOutMsg, err := proto.Marshal(syncOut)
	if err != nil {
		fmt.Println("[离线消息] proto.Marshal error, err:", err)
		return
	}
	ackBytes, err := GetOutputMsg(pb.CmdType_CT_Sync, int32(common.OK), syncOutMsg)
	if err != nil {
		fmt.Println("[离线消息] proto.Marshal err:", err)
		return
	}
	// 回复
	r.conn.SendMsg(r.conn.GetUserId(), ackBytes)
}
