package main

import (
	"GoChat/pkg/protocol/pb"
	"GoChat/pkg/util"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ResendCountMax = 3 // 超时重传最大次数
)

type Client struct {
	conn                 *websocket.Conn
	token                string
	userId               uint64
	clientId             uint64
	clientId2Cancel      map[uint64]context.CancelFunc // clientId 到 context 的映射
	clientId2CancelMutex sync.Mutex
	seq                  uint64 // 本地消息最大同步序列号

	sendCh chan []byte // 写入
}

func NewClient(userId, token, host string) *Client {
	// 创建 client
	c := &Client{
		clientId2Cancel: make(map[uint64]context.CancelFunc),
		token:           token,
		userId:          util.StrToUint64(userId),
		sendCh:          make(chan []byte, 1024),
	}

	// 连接 websocket
	conn, _, err := websocket.DefaultDialer.Dial(host+"/ws", http.Header{})
	if err != nil {
		panic(err)
	}
	c.conn = conn
	// 向 websocket 发送登录请求
	c.login()

	// 心跳
	go c.heartbeat()

	// 写
	go c.write()

	// 读
	go c.read()

	return c
}

func (c *Client) read() {
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		msg := new(pb.Output)
		err = proto.Unmarshal(bytes, msg)
		if err != nil {
			panic(err)
		}

		// 只收两种，Message 收取下行消息和 ACK，上行消息ACK回复
		switch msg.Type {
		case pb.CmdType_CT_Message:
			// 计算接收消息数量
			atomic.AddInt64(&receivedMessageCount, 1)
			msgTimer.updateEndTime()

			pushMsg := new(pb.PushMsg)
			err = proto.Unmarshal(msg.Data, pushMsg)
			if err != nil {
				panic(err)
			}
			// 更新 seq
			seq := pushMsg.Msg.Seq
			if c.seq < seq {
				c.seq = seq
			}
		case pb.CmdType_CT_ACK: // 收到 ACK
			ackMsg := new(pb.ACKMsg)
			err = proto.Unmarshal(msg.Data, ackMsg)
			if err != nil {
				panic(err)
			}

			switch ackMsg.Type {
			case pb.ACKType_AT_Up: // 收到上行消息的 ACK
				// 计算接收消息数量
				atomic.AddInt64(&receivedMessageCount, 1)
				msgTimer.updateEndTime()

				// 取消超时重传
				clientId := ackMsg.ClientId
				c.clientId2CancelMutex.Lock()
				if cancel, ok := c.clientId2Cancel[clientId]; ok {
					// 取消超时重传
					cancel()
					delete(c.clientId2Cancel, clientId)
					//fmt.Println("取消超时重传，clientId:", clientId)
				}
				c.clientId2CancelMutex.Unlock()
				// 更新客户端本地维护的 seq
				seq := ackMsg.Seq
				if c.seq < seq {
					c.seq = seq
				}
			}
		default:
			fmt.Println("未知消息类型")
		}
	}
}

func (c *Client) write() {
	for {
		select {
		case bytes, ok := <-c.sendCh:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
				return
			}
		}
	}
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Minute * 2)
	for range ticker.C {
		c.sendMsg(pb.CmdType_CT_Heartbeat, &pb.HeartbeatMsg{})
	}
}

func (c *Client) login() {
	c.sendMsg(pb.CmdType_CT_Login, &pb.LoginMsg{
		Token: []byte(c.token),
	})
}

// send 发送消息，启动超时重试
func (c *Client) send(chatId int64) {
	message := &pb.Message{
		SessionType: pb.SessionType_ST_Group,                       // 群聊
		ReceiverId:  uint64(chatId),                                // 发送到该群
		SenderId:    c.userId,                                      // 发送者
		MessageType: pb.MessageType_MT_Text,                        // 文本
		Content:     []byte("文本聊天消息" + util.Uint64ToStr(c.userId)), // 消息
		SendTime:    time.Now().UnixMilli(),                        // 发送时间
	}
	UpMsg := &pb.UpMsg{
		Msg:      message,
		ClientId: c.getClientId(),
	}
	// 发送消息
	c.sendMsg(pb.CmdType_CT_Message, UpMsg)

	// 启动超时重传
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		maxRetry := ResendCountMax // 最大重试次数
		retryCount := 0
		retryInterval := time.Millisecond * 500 // 重试间隔
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
				if retryCount >= maxRetry {
					return
				}
				c.sendMsg(pb.CmdType_CT_Message, UpMsg)
				retryCount++
			}
		}
	}(ctx)

	c.clientId2CancelMutex.Lock()
	c.clientId2Cancel[UpMsg.ClientId] = cancel
	c.clientId2CancelMutex.Unlock()

}

func (c *Client) getClientId() uint64 {
	c.clientId++
	return c.clientId
}

//  客户端向服务端发送上行消息
func (c *Client) sendMsg(cmdType pb.CmdType, msg proto.Message) {
	// 组装顶层数据
	cmdMsg := &pb.Input{
		Type: cmdType,
		Data: nil,
	}
	if msg != nil {
		data, err := proto.Marshal(msg)
		if err != nil {
			panic(err)
		}
		cmdMsg.Data = data
	}

	bytes, err := proto.Marshal(cmdMsg)
	if err != nil {
		panic(err)
	}

	// 发送
	c.sendCh <- bytes
}
