package model

import (
	"GoChat/pkg/db"
	"time"
)

type Message struct {
	ID          uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	UserID      uint64    `gorm:"not null;comment:'用户id，发送者id'" json:"user_id"`
	SessionType int8      `gorm:"not null;comment:'聊天类型，群聊/单聊'" json:"session_type"`
	ReceiverId  uint64    `gorm:"not null;comment:'接收者id，群聊id/用户id'" json:"receiver_id"`
	MessageType int8      `gorm:"not null;comment:'消息类型,语言、文字、图片'" json:"message_type"`
	Content     string    `gorm:"not null;comment:'消息内容'" json:"content"`
	CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (*Message) TableName() string {
	return "message"
}

func CreateMessage(msg *Message) error {
	return db.DB.Create(msg).Error
}
