package main

import (
	"GoChat/pkg/protocol/pb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	conn   *websocket.Conn
	token  string
	userId uint64
}

// websocket 客户端
func main() {
	// http 登录，获取 token
	client := Login()

	// 连接 websocket 服务
	client.Start()

	select {}
}

func (c *Client) Start() {
	// 连接 websocket
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:9091/ws", http.Header{})
	if err != nil {
		panic(err)
	}
	c.conn = conn

	fmt.Println("与 websocket 建立连接")
	time.Sleep(time.Second)
	// 向 websocket 发送登录请求
	c.Login()

	// 心跳
	go c.Heartbeat()

	// 收取消息
	c.Receive()
}

func (c *Client) Heartbeat() {
	//  2min 一次
	ticker := time.NewTicker(time.Second * 2 * 60)
	for range ticker.C {
		c.SendMsg(pb.CmdType_Heartbeat, nil)
		fmt.Println("发送心跳", time.Now().Format("2006-01-02 15:04:05"))
	}
}

func (c *Client) Receive() {
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}
		c.HandlerMessage(bytes)
	}
}

// HandlerMessage 客户端消息处理
func (c *Client) HandlerMessage(bytes []byte) {
	var msg pb.CmdMsg
	err := proto.Unmarshal(bytes, &msg)
	if err != nil {
		panic(err)
	}

	switch msg.Type {
	case pb.CmdType_Login: // 登录
		fmt.Println("收到登录ACK", time.Now().Format("2006-01-02 15:04:05"))
	case pb.CmdType_Heartbeat: // 心跳
		fmt.Println("收到心跳ACK", time.Now().Format("2006-01-02 15:04:05"))
	case pb.CmdType_SYNC: // 离线消息同步

	case pb.CmdType_ACK: // 消息ACK

	case pb.CmdType_Up: // 上行消息

	case pb.CmdType_Push: // 服务端不可能收到服务端发送的下行消息
		fmt.Println("消息类型错误")
	default:
		fmt.Println("未知消息类型")
	}
}

// Login websocket 登录
func (c *Client) Login() {
	fmt.Println("websocket login...")
	// 组装底层数据
	loginMsg := &pb.LoginMsg{
		UserId: c.userId,
		Token:  []byte(c.token),
	}
	c.SendMsg(pb.CmdType_Login, loginMsg)
}

// SendMsg 客户端向服务端发送上行消息
func (c *Client) SendMsg(cmdType pb.CmdType, msg proto.Message) {
	// 组装顶层数据
	cmdMsg := &pb.CmdMsg{
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
	err = c.conn.WriteMessage(websocket.BinaryMessage, bytes)
	if err != nil {
		panic(err)
	}
}

// Login 用户http登录获取 token
func Login() *Client {
	// 读取 phone_number 和 password 参数
	var phoneNumber, password string
	fmt.Print("Enter phone_number: ")
	fmt.Scanln(&phoneNumber)
	fmt.Print("Enter password: ")
	fmt.Scanln(&password)

	// 创建一个 url.Values 对象，并将 phone_number 和 password 参数添加到其中
	data := url.Values{}
	data.Set("phone_number", phoneNumber)
	data.Set("password", password)

	// 向服务器发送 POST 请求
	resp, err := http.PostForm("http://localhost:9090/login", data)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		panic(err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status code: %d\n", resp.StatusCode)
		panic(err)
	}

	// 读取返回数据
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// 获取 token
	var p fastjson.Parser
	v, err := p.Parse(string(bodyData))
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		panic(err)
	}

	client := &Client{}

	code, _ := v.Get("code").Int()
	msg := v.Get("msg").String()

	if code != 200 {
		panic(fmt.Sprintf("登录失败, %s", msg))
	}

	respData := v.Get("data")

	client.token = string(respData.GetStringBytes("token"))
	clientStr := string(respData.GetStringBytes("user_id"))
	client.userId, err = strconv.ParseUint(clientStr, 10, 64)
	if err != nil {
		panic(err)
	}

	fmt.Println("client code:", code, " ", msg)
	fmt.Println("token:", client.token, "userId", client.userId)
	return client
}
