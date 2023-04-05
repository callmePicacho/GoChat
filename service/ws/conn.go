package ws

import (
	"GoChat/config"
	"GoChat/lib/cache"
	"GoChat/pkg/protocol/pb"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

// Conn 连接实例
// 1. 启动读写线程
// 2. 读线程读到数据后，根据数据类型获取处理函数，交给 worker 队列调度执行
type Conn struct {
	ConnId           uint64          // 连接编号，通过对编号取余，能够让 Conn 始终进入同一个 worker，保持有序性
	server           *Server         // 当前连接属于哪个 server
	UserId           uint64          // 连接所属用户id
	UserIdMutex      sync.RWMutex    // 保护 userId 的锁
	Socket           *websocket.Conn // 用户连接
	sendCh           chan []byte     // 用户要发送的数据
	isClose          bool            // 连接状态
	isCloseMutex     sync.RWMutex    // 保护 isClose 的锁
	exitCh           chan struct{}   // 通知 writer 退出
	maxClientId      uint64          // 该连接收到的最大 clientId，确保消息的可靠性
	maxClientIdMutex sync.Mutex      // 保护 maxClientId 的锁

	lastHeartBeatTime time.Time  // 最后活跃时间
	heartMutex        sync.Mutex // 保护最后活跃时间的锁
}

func NewConnection(server *Server, wsConn *websocket.Conn, ConnId uint64) *Conn {
	return &Conn{
		ConnId:            ConnId,
		server:            server,
		UserId:            0, // 此时用户未登录， userID 为 0
		Socket:            wsConn,
		sendCh:            make(chan []byte, 1),
		isClose:           false,
		exitCh:            make(chan struct{}, 1),
		lastHeartBeatTime: time.Now(), // 刚连接时初始化，避免正好遇到清理执行，如果连接没有后续操作，将会在下次被心跳检测踢出
	}
}

func (c *Conn) Start() {
	// 开启从客户端读取数据流程的 goroutine
	go c.StartReader()

	// 开启用于写回客户端数据流程的 goroutine
	go c.StartWriter()
}

// StartReader 用于从客户端中读取数据
func (c *Conn) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr(), "[conn Reader exit!]")
	defer c.Stop()

	for {
		// 阻塞读
		_, data, err := c.Socket.ReadMessage()
		if err != nil {
			fmt.Println("read msg data error ", err)
			return
		}

		// 消息处理
		c.HandlerMessage(data)
	}
}

// HandlerMessage 消息处理
func (c *Conn) HandlerMessage(bytes []byte) {
	// TODO 所有错误都需要写回给客户端
	// 消息解析 proto string -> struct
	input := new(pb.Input)
	err := proto.Unmarshal(bytes, input)
	if err != nil {
		fmt.Println("unmarshal error", err)
		return
	}
	//fmt.Println("收到消息：", input)

	// 对未登录用户进行拦截
	if input.Type != pb.CmdType_CT_Login && c.GetUserId() == 0 {
		return
	}

	req := &Req{
		conn: c,
		data: input.Data,
		f:    nil,
	}

	switch input.Type {
	case pb.CmdType_CT_Login: // 登录
		req.f = req.Login
	case pb.CmdType_CT_Heartbeat: // 心跳
		req.f = req.Heartbeat
	case pb.CmdType_CT_Message: // 上行消息
		req.f = req.MessageHandler
	case pb.CmdType_CT_ACK: // ACK TODO

	case pb.CmdType_CT_Sync: // 离线消息同步
		req.f = req.Sync
	default:
		fmt.Println("未知消息类型")
	}

	if req.f == nil {
		return
	}

	// 更新心跳时间
	c.KeepLive()

	// 送入worker队列等待调度执行
	c.server.SendMsgToTaskQueue(req)
}

// SendMsg 根据 userId 给相应 socket 发送消息
func (c *Conn) SendMsg(userId uint64, bytes []byte) {
	c.isCloseMutex.RLock()
	defer c.isCloseMutex.RUnlock()

	// 已关闭
	if c.isClose {
		fmt.Println("connection closed when send msg")
		return
	}

	// 根据 userId 找到对应 socket
	conn := c.server.GetConn(userId)
	if conn == nil {
		return
	}

	// 发送
	conn.sendCh <- bytes

	return
}

// StartWriter 向客户端写数据
func (c *Conn) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr(), "[conn Writer exit!]")

	var err error
	for {
		select {
		case data := <-c.sendCh:
			if err = c.Socket.WriteMessage(websocket.BinaryMessage, data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			// 更新心跳时间
			c.KeepLive()
		case <-c.exitCh:
			return
		}
	}
}

func (c *Conn) Stop() {
	c.isCloseMutex.Lock()
	defer c.isCloseMutex.Unlock()

	if c.isClose {
		return
	}

	// 关闭 socket 连接
	_ = c.Socket.Close()
	// 关闭 writer
	c.exitCh <- struct{}{}

	if c.GetUserId() != 0 {
		// 将连接从connMap中移除
		c.server.RemoveConn(c.GetUserId())
		// 用户下线
		_ = cache.DelUserOnline(c.GetUserId())
	}

	c.isClose = true

	// 关闭管道
	close(c.exitCh)
	close(c.sendCh)

	fmt.Println("Conn Stop() ... UserId = ", c.GetUserId())
}

// KeepLive 更新心跳
func (c *Conn) KeepLive() {
	now := time.Now()
	c.heartMutex.Lock()
	defer c.heartMutex.Unlock()

	c.lastHeartBeatTime = now
}

// IsAlive 是否存活
func (c *Conn) IsAlive() bool {
	now := time.Now()

	c.heartMutex.Lock()
	c.isCloseMutex.RLock()
	defer c.isCloseMutex.RUnlock()
	defer c.heartMutex.Unlock()

	if c.isClose || now.Sub(c.lastHeartBeatTime) > time.Duration(config.GlobalConfig.App.HeartbeatTimeout)*time.Second {
		return false
	}
	return true
}

// GetUserId 获取 userId
func (c *Conn) GetUserId() uint64 {
	c.UserIdMutex.RLock()
	defer c.UserIdMutex.RUnlock()

	return c.UserId
}

// SetUserId 设置 UserId
func (c *Conn) SetUserId(userId uint64) {
	c.UserIdMutex.Lock()
	defer c.UserIdMutex.Unlock()

	c.UserId = userId
}

func (c *Conn) CompareAndIncrClientID(newMaxClientId uint64) bool {
	c.maxClientIdMutex.Lock()
	defer c.maxClientIdMutex.Unlock()

	//fmt.Println("收到的 newMaxClientId 是：", newMaxClientId, "此时 c.maxClientId 是：", c.maxClientId)
	if c.maxClientId+1 == newMaxClientId {
		c.maxClientId++
		return true
	}
	return false
}

// RemoteAddr 获取远程客户端地址
func (c *Conn) RemoteAddr() string {
	return c.Socket.RemoteAddr().String()
}
