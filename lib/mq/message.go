package mq

import (
	"GoChat/model"
	"GoChat/pkg/mq"
	"fmt"
	"github.com/wagslane/go-rabbitmq"
)

const (
	MessageQueue        = "message.queue"
	MessageRoutingKey   = "message.routing.key"
	MessageExchangeName = "message.exchange.name"
)

var (
	MessageMQ *mq.Conn
)

func InitMessageMQ(url string) {
	MessageMQ = mq.InitRabbitMQ(url, MessageCreateHandler, MessageQueue, MessageRoutingKey, MessageExchangeName)
}

func MessageCreateHandler(d rabbitmq.Delivery) rabbitmq.Action {
	messageModels := model.JsonToMessage(d.Body)
	err := model.CreateMessage(messageModels...)
	if err != nil {
		fmt.Println("[MessageCreateHandler] model.CreateMessage 失败，err:", err)
		return rabbitmq.NackDiscard
	}

	//fmt.Println("处理完消息：", string(d.Body))
	return rabbitmq.Ack
}
