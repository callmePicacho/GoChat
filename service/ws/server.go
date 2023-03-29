package ws

import (
	"GoChat/config"
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

// Server 连接管理
// 1. 连接管理
// 2. 工作队列
type Server struct {
	IP                string
	Port              string
	connUnLoginMap    map[*Conn]struct{} // 还未验证连接，调用 /ws 的 websocket 连接
	connUnLoginWMutex sync.RWMutex       // 还未验证连接读写锁
	connMap           map[uint64]*Conn   // 验证通过的用户连接，通过验证将会把 conn 从 connUnLoginMap 移到 connMap 中 k-用户userid v-连接
	connRWMutex       sync.RWMutex       // 验证通过连接读写锁
	taskQueue         []chan *Req        // 工作池
}

func NewServer() *Server {
	return &Server{
		IP:             config.GlobalConfig.App.IP,
		Port:           config.GlobalConfig.App.WebsocketPort,
		connUnLoginMap: make(map[*Conn]struct{}, 10000),                           // 提前预留大小
		connMap:        make(map[uint64]*Conn, 10000),                             // 提前预留大小
		taskQueue:      make([]chan *Req, config.GlobalConfig.App.WorkerPoolSize), // 初始worker队列中，worker个数
	}
}

// AddConnUnLogin 添加连接
func (cm *Server) AddConnUnLogin(conn *Conn) {
	cm.connUnLoginWMutex.Lock()
	defer cm.connUnLoginWMutex.Unlock()

	cm.connUnLoginMap[conn] = struct{}{}
	fmt.Printf("connection add to Server successfully: connUnLogin num = %d \n", len(cm.connUnLoginMap))
}

// RemoveConnUnLogin 删除连接
func (cm *Server) RemoveConnUnLogin(conn *Conn) {
	cm.connUnLoginWMutex.Lock()
	defer cm.connUnLoginWMutex.Unlock()

	delete(cm.connUnLoginMap, conn)

	fmt.Println("connection Remove UserID=", conn.UserId, " successfully: connUnLogin num = ", len(cm.connUnLoginMap))
}

// InConnUnLogin 判断 conn 是否存在 unlogin map 中
func (cm *Server) InConnUnLogin(conn *Conn) bool {
	cm.connUnLoginWMutex.RLock()
	defer cm.connUnLoginWMutex.RUnlock()

	_, ok := cm.connUnLoginMap[conn]

	return ok
}

// GetConnUnLoginAll 获取全部未验证的连接
func (cm *Server) GetConnUnLoginAll() []*Conn {
	cm.connUnLoginWMutex.RLock()
	defer cm.connUnLoginWMutex.RUnlock()

	conns := make([]*Conn, 0, len(cm.connUnLoginMap))
	for conn, _ := range cm.connUnLoginMap {
		conns = append(conns, conn)
	}
	return conns
}

// AddConn 添加连接
func (cm *Server) AddConn(conn *Conn) {
	cm.connRWMutex.Lock()
	defer cm.connRWMutex.Unlock()

	cm.connMap[conn.UserId] = conn
	fmt.Printf("connection UserId=%d add to Server connMap successfully: conn num = %d \n", conn.UserId, len(cm.connMap))
}

// RemoveConn 删除连接
func (cm *Server) RemoveConn(conn *Conn) {
	cm.connRWMutex.Lock()
	defer cm.connRWMutex.Unlock()

	delete(cm.connMap, conn.UserId)

	fmt.Println("connection Remove UserID=", conn.UserId, " successfully: conn num = ", len(cm.connMap))
}

// GetConn 根据userid获取相应的连接
func (cm *Server) GetConn(userId uint64) (*Conn, error) {
	cm.connRWMutex.RLock()
	defer cm.connRWMutex.RUnlock()

	if conn, ok := cm.connMap[userId]; ok {
		return conn, nil
	}

	return nil, errors.New("connection not found")
}

// GetConnAll 获取全部连接
func (cm *Server) GetConnAll() []*Conn {
	cm.connRWMutex.RLock()
	defer cm.connRWMutex.RUnlock()

	conns := make([]*Conn, 0, len(cm.connMap))
	for _, conn := range cm.connMap {
		conns = append(conns, conn)
	}
	return conns
}

// StartWorkerPool 启动 worker 工作池
func (cm *Server) StartWorkerPool() {
	// 初始化并启动 worker 工作池
	for i := 0; i < len(cm.taskQueue); i++ {
		// 初始化
		cm.taskQueue[i] = make(chan *Req, config.GlobalConfig.App.MaxWorkerTask) // 初始化worker队列中，每个worker的队列长度
		// 启动
		go cm.StartOneWorker(i, cm.taskQueue[i])
	}
}

// StartOneWorker 启动 worker 的工作流程
func (cm *Server) StartOneWorker(workerID int, taskQueue chan *Req) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	for {
		select {
		case req := <-taskQueue:
			req.f()
		}
	}
}

// SendMsgToTaskQueue 将消息交给 taskQueue，由 worker 调度处理
func (cm *Server) SendMsgToTaskQueue(req *Req) {
	if len(cm.taskQueue) > 0 {
		// hash
		i := rand.Intn(len(cm.taskQueue))

		// 将消息发给对应的 taskQueue
		cm.taskQueue[i] <- req
	} else {
		go req.f()
	}
}

// Addr 获取网关地址
func (cm *Server) Addr() string {
	return fmt.Sprintf("%s:%s", cm.IP, cm.Port)
}
