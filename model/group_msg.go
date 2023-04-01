package model

import (
	"GoChat/pkg/db"
	"time"
)

type GroupMsg struct {
	ID          uint64    `gorm:"column:id;primary_key;auto_increment;comment:'自增主键'" json:"id"`
	UserID      uint64    `gorm:"column:user_id;not null;comment:'用户id，发送者id'" json:"user_id"`
	GroupID     uint64    `gorm:"column:group_id;not null;comment:'群组id'" json:"group_id"`
	MessageType int8      `gorm:"column:message_type;not null;comment:'消息类型,语言、文字、图片'" json:"message_type"`
	SendTime    time.Time `gorm:"column:send_time;not null;comment:'消息发送时间'" json:"send_time"`
	Content     []byte    `gorm:"column:content;not null;comment:'消息内容'" json:"content"`
	CreateTime  time.Time `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (m *GroupMsg) TableName() string {
	return "group_msg"
}

func CreateGroupMsg(groupUser *GroupMsg) error {
	return db.DB.Create(groupUser).Error
}
