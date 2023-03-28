package model

import (
	"GoChat/pkg/db"
	"time"
)

type Friend struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	UserID     uint64    `gorm:"not null;comment:'用户id'" json:"user_id"`
	FriendID   uint64    `gorm:"not null;comment:'好友id'" json:"friend_id"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (*Friend) TableName() string {
	return "friend"
}

// IsFriend 查询是否为好友关系
func IsFriend(userId, friendId uint64) (bool, error) {
	var cnt int64
	err := db.DB.Model(&Friend{}).Where("user_id = ? and friend_id = ?", userId, friendId).
		Or("friend_id = ? and user_id = ?", userId, friendId). // 反查
		Count(&cnt).Error
	return cnt > 0, err
}

func CreateFriend(friend *Friend) error {
	return db.DB.Create(friend).Error
}
