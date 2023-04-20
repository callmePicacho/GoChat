package model

import "time"

type UID struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	BusinessID string    `gorm:"not null;uniqueIndex:uk_business_id;comment:'业务id'" json:"business_id"`
	MaxID      uint64    `gorm:"default:NULL;comment:'最大id'" json:"max_id"`
	Step       int       `gorm:"default:NULL;comment:'步长'" json:"step"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (u *UID) TableName() string {
	return "uid"
}
