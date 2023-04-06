package mq

import (
	"fmt"
	"github.com/wagslane/go-rabbitmq"
)

type Conn struct {
	conn         *rabbitmq.Conn
	consumer     *rabbitmq.Consumer
	publisher    *rabbitmq.Publisher
	queue        string
	routingKey   string
	exchangeName string
}

// InitRabbitMQ 初始化连接
// 启动消费者、初始化生产者
func InitRabbitMQ(url string, f rabbitmq.Handler, queue, routingKey, exchangeName string) *Conn {
	// 初始化连接
	conn, err := rabbitmq.NewConn(url)
	if err != nil {
		panic(err)
	}
	// 消费者，注册时已经启动了
	consumer, err := rabbitmq.NewConsumer(
		conn,
		f,     // 实际进行消费处理的函数
		queue, // 队列名称
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),     // routing-key
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName), // exchange 名称
		rabbitmq.WithConsumerOptionsExchangeDeclare,            // 声明交换器
	)
	if err != nil {
		panic(err)
	}

	// 生产者
	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName), // exchange 名称
		rabbitmq.WithPublisherOptionsExchangeDeclare,            // 声明交换器
	)
	if err != nil {
		panic(err)
	}

	// 连接被拒绝
	publisher.NotifyReturn(func(r rabbitmq.Return) {
		//log.Printf("message returned from server: %s", string(r.Body))
	})

	// 提交确认
	publisher.NotifyPublish(func(c rabbitmq.Confirmation) {
		//log.Printf("message confirmed from server. tag: %v, ack: %v", c.DeliveryTag, c.Ack)
	})

	return &Conn{
		conn:         conn,
		consumer:     consumer,
		publisher:    publisher,
		queue:        queue,
		routingKey:   routingKey,
		exchangeName: exchangeName,
	}
}

// Publish 发送消息，该消息实际由执行 InitRabbitMQ 注册时传入的 f 消费
func (c *Conn) Publish(data []byte) error {
	if data == nil || len(data) == 0 {
		fmt.Println("data 为空，publish 不发送")
		return nil
	}
	return c.publisher.Publish(
		data,
		[]string{c.routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsPersistentDelivery,       // 消息持久化
		rabbitmq.WithPublishOptionsExchange(c.exchangeName), // 要发送的 exchange
	)
}

func (c *Conn) Close() {
	c.conn.Close()
	c.consumer.Close()
	c.publisher.Close()
}
