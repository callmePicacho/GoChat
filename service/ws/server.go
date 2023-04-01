package ws

import (
	"GoChat/config"
	"fmt"
	"sync"
)

var (
	ConnManager *Server
	once        sync.Once
)

// Server 连接管理
// 1. 连接管理
// 2. 工作队列
type Server struct {
	connMap   sync.Map    // 登录的用户连接 k-用户userid v-连接
	taskQueue []chan *Req // 工作池
}

func GetServer() *Server {
	once.Do(func() {
		ConnManager = &Server{
			taskQueue: make([]chan *Req, config.GlobalConfig.App.WorkerPoolSize), // 初始worker队列中，worker个数
		}
	})
	return ConnManager
}

// Stop 关闭服务
func (cm *Server) Stop() {
	fmt.Println("server stop ...")
	var wg sync.WaitGroup
	connAll := cm.GetConnAll()
	for _, conn := range connAll {
		wg.Add(1)
		conn := conn
		go func() {
			defer wg.Done()
			conn.Stop()
		}()
	}
	wg.Wait()
}

// AddConn 添加连接
func (cm *Server) AddConn(userId uint64, conn *Conn) {
	cm.connMap.Store(userId, conn)
	fmt.Printf("connection UserId=%d add to Server\n", userId)
}

// RemoveConn 删除连接
func (cm *Server) RemoveConn(userId uint64) {
	cm.connMap.Delete(userId)
	fmt.Printf("connection UserId=%d remove from Server\n", userId)
}

// GetConn 根据userid获取相应的连接
func (cm *Server) GetConn(userId uint64) *Conn {
	value, ok := cm.connMap.Load(userId)
	if ok {
		return value.(*Conn)
	}
	return nil
}

// GetConnAll 获取全部连接
func (cm *Server) GetConnAll() []*Conn {
	conns := make([]*Conn, 0)
	cm.connMap.Range(func(key, value interface{}) bool {
		conn := value.(*Conn)
		conns = append(conns, conn)
		return true
	})
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
		// 根据ConnID来分配当前的连接应该由哪个worker负责处理
		// 轮询的平均分配法则

		//得到需要处理此条连接的workerID
		workerID := req.conn.ConnId % uint64(len(cm.taskQueue))

		// 将消息发给对应的 taskQueue
		cm.taskQueue[workerID] <- req
	} else {
		go req.f()
	}
}
