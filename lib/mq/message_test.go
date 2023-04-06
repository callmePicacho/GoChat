package mq

import (
	"GoChat/model"
	"encoding/json"
	"testing"
	"time"
)

func TestMessageMQ(t *testing.T) {
	url := "amqp://guest:guest@localhost:5672"
	InitMessageMQ(url)

	msgs := make([]*model.Message, 0)
	msgs = append(msgs, &model.Message{
		ID:          1,
		UserID:      2,
		SenderID:    3,
		SessionType: 4,
		ReceiverId:  5,
		MessageType: 6,
		Content:     []byte("我去"),
		Seq:         7,
		SendTime:    time.Now(),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	})
	data, err := json.Marshal(msgs)
	if err != nil {
		panic(err)
	}
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			err = MessageMQ.Publish(data)
			if err != nil {
				panic(err)
			}
		}
	}
}
