package model

import (
	"GoChat/pkg/db"
	"GoChat/pkg/protocol/pb"
	"encoding/json"
	"fmt"
	"time"
)

const MessageLimit = 50 // 最大消息同步数量

// Message 单聊消息
type Message struct {
	ID          uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	UserID      uint64    `gorm:"not null;comment:'用户id，指接受者用户id'" json:"user_id"`
	SenderID    uint64    `gorm:"not null;comment:'发送者用户id'"`
	SessionType int8      `gorm:"not null;comment:'聊天类型，群聊/单聊'" json:"session_type"`
	ReceiverId  uint64    `gorm:"not null;comment:'接收者id，群聊id/用户id'" json:"receiver_id"`
	MessageType int8      `gorm:"not null;comment:'消息类型,语言、文字、图片'" json:"message_type"`
	Content     []byte    `gorm:"not null;comment:'消息内容'" json:"content"`
	Seq         uint64    `gorm:"not null;comment:'消息序列号'" json:"seq"`
	SendTime    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'消息发送时间'" json:"send_time"`
	CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (*Message) TableName() string {
	return "message"
}

func MessagesToJson(messages ...*Message) []byte {
	if len(messages) == 0 {
		return nil
	}
	bytes, err := json.Marshal(&messages)
	if err != nil {
		fmt.Println("json.Marshal(messages) 失败,err:", err)
		return nil
	}
	return bytes
}

func JsonToMessage(bytes []byte) []*Message {
	messages := make([]*Message, 0)
	err := json.Unmarshal(bytes, &messages)
	if err != nil {
		fmt.Println("json.Unmarshal(bytes, message) 失败,err:", err)
		return nil
	}
	return messages
}

func MessagesToPB(messages []Message) []*pb.Message {
	pbMessages := make([]*pb.Message, 0, len(messages))
	for _, message := range messages {
		pbMessages = append(pbMessages, &pb.Message{
			SessionType: pb.SessionType(message.SessionType),
			ReceiverId:  message.ReceiverId,
			SenderId:    message.SenderID,
			MessageType: pb.MessageType(message.MessageType),
			Content:     message.Content,
			Seq:         message.Seq,
		})
	}
	return pbMessages
}

func CreateMessage(msgs ...*Message) error {
	return db.DB.Create(msgs).Error
}

func ListByUserIdAndSeq(userId, seq uint64, limit int) ([]Message, bool, error) {
	var cnt int64
	err := db.DB.Model(&Message{}).Where("user_id = ? and seq > ?", userId, seq).
		Count(&cnt).Error
	if err != nil {
		return nil, false, err
	}
	if cnt == 0 {
		return nil, false, nil
	}

	var messages []Message
	err = db.DB.Model(&Message{}).Where("user_id = ? and seq > ?", userId, seq).
		Limit(limit).Order("seq ASC").Find(&messages).Error
	if err != nil {
		return nil, false, err
	}
	return messages, cnt > int64(limit), nil
}
