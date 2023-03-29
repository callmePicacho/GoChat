package ws

import (
	"GoChat/lib/cache"
	"GoChat/pkg/protocol/pb"
	"GoChat/pkg/util"
	"fmt"
	"github.com/golang/protobuf/proto"
	"sync"
)

// Handler 路由函数
type Handler func()

// Req 请求
type Req struct {
	conn       *Conn   // 连接
	data       []byte  // 客户端发送的请求数据
	f          Handler // 该请求需要执行的路由函数
	LoginMutex sync.Mutex
}

func (r *Req) Login() {
	r.LoginMutex.Lock()
	defer r.LoginMutex.Unlock()

	// 检查用户是否已登录  TODO 还需要验证 Redis 中是否存在用户数据并做相应处理
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

	// Redis 存储用户数据 k: userId,  v: 网关地址
	userId := r.conn.GetUserId()
	serverAddr := r.conn.server.Addr()
	err = cache.SetUserOnline(userId, serverAddr)
	if err != nil {
		fmt.Println("[用户登录] 系统错误")
		return
	}

	// 转换 user_id 所在 map
	if r.conn.server.InConnUnLogin(r.conn) {
		// 从 connUnLogin 中删除
		r.conn.server.RemoveConnUnLogin(r.conn)
		// 加入到 connMap 中
		r.conn.server.AddConn(r.conn)
	}

	// 回复ACK
	msg := &pb.CmdMsg{
		Type: pb.CmdType_Login,
		Data: nil,
	}
	bytes, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("[用户登录] proto.Marshal err:", err)
		return
	}

	r.conn.SendMsg(loginMsg.UserId, bytes)
}

func (r *Req) HeartBeat() {

}
