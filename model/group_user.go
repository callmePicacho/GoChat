package model

import (
	"GoChat/pkg/db"
	"time"
)

type GroupUser struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	GroupID    uint64    `gorm:"not null;comment:'组id'" json:"group_id"`
	UserID     uint64    `gorm:"not null;comment:'用户id'" json:"user_id"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (*GroupUser) TableName() string {
	return "group_user"
}

// IsBelongToGroup 验证用户是否属于群
func IsBelongToGroup(userId, groupId uint64) (bool, error) {
	var cnt int64
	err := db.DB.Model(&GroupUser{}).
		Where("user_id = ? and group_id = ?", userId, groupId).
		Count(&cnt).Error
	return cnt > 0, err
}

func GetGroupUserIdsByGroupId(groupId uint64) ([]uint64, error) {
	var ids []uint64
	err := db.DB.Model(&GroupUser{}).Where("group_id = ?", groupId).Pluck("user_id", &ids).Error
	return ids, err
}
