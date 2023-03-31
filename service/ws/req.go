package ws

import (
	"GoChat/common"
	"GoChat/config"
	"GoChat/lib/cache"
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
	// TODO 要防止多个连接使用相同 user_id + token 进行 Login，还需要验证 Redis 中是否存在用户数据并做相应处理
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
	if userClaims.UserId != loginMsg.UserId {
		fmt.Println("[用户登录] 非法用户")
		return
	}

	// 设置 user_id
	r.conn.SetUserId(loginMsg.UserId)

	// Redis 存储用户数据 k: userId,  v: grpc地址，方便用户能直接通过这个地址进行 rpc 方法调用
	grpcServerAddr := fmt.Sprintf("%s:%s", config.GlobalConfig.App.IP, config.GlobalConfig.App.RPCPort)
	err = cache.SetUserOnline(loginMsg.UserId, grpcServerAddr)
	if err != nil {
		fmt.Println("[用户登录] 系统错误")
		return
	}

	// 回复ACK
	bytes, err := GetOutputMsg(pb.CmdType_CT_Login, int32(common.OK), nil)
	if err != nil {
		fmt.Println("[用户登录] proto.Marshal err:", err)
		return
	}

	r.conn.SendMsg(loginMsg.UserId, bytes)

	// 加入到 connMap 中
	r.conn.server.AddConn(loginMsg.UserId, r.conn)
}

func (r *Req) HeartBeat() {
	// 回复心跳
	bytes, err := GetOutputMsg(pb.CmdType_CT_Heartbeat, int32(common.OK), nil)
	if err != nil {
		fmt.Println("[心跳] proto.Marshal err:", err)
		return
	}
	r.conn.SendMsg(r.conn.GetUserId(), bytes)
}

// MessageHandler 消息处理，处理客户端发送给服务端的消息
// A 发送消息给服务端，服务端收到消息处理后发给 B
// 包括：单聊、群聊
func (r *Req) MessageHandler() {
	// 消息解析 proto string -> struct
	msg := new(pb.Message)
	err := proto.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[消息处理] unmarshal error,err:", err)
		return
	}

	if msg.SenderId != r.conn.GetUserId() {
		fmt.Println("[消息处理] 发送者有误")
		return
	}

	if msg.ReceiverId == r.conn.GetUserId() {
		fmt.Println("[消息处理] 接收者有误")
		return
	}

	// 得到要转发给 B 的消息
	bytes, err := GetOutputMsg(pb.CmdType_CT_Message, int32(common.OK), r.data)
	if err != nil {
		fmt.Println("[消息处理] Marshal error,err:", err)
		return
	}

	// 单聊、群聊（写扩散）
	switch msg.SessionType {
	case pb.SessionType_ST_Single:
		// 消息本身 和 要发送的消息
		err = SendToUser(msg, bytes)
	case pb.SessionType_ST_Group:
		err = SendToGroup(msg, bytes)
	default:
		fmt.Println("[消息处理] 会话类型错误")
		return
	}

	// TODO 组装回复 A 的 ACK
}
