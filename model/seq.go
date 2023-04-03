package model

import (
	"GoChat/pkg/db"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Seq struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	ObjectType int8      `gorm:"not null;comment:'对象类型,1:用户；2：群组'" json:"object_type"`
	ObjectID   uint64    `gorm:"not null;comment:'对象id'" json:"object_id"`
	Seq        uint64    `gorm:"not null;comment:'消息序列号'" json:"seq"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (s *Seq) TableName() string {
	return "seq"
}

const (
	SeqObjectTypeUser = 1 // 用户
)

func Incr(objectType int8, objectId uint64) (uint64, error) {
	var seq uint64
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 查询并锁定
		err := tx.Model(&Seq{}).Set("gorm:query_option", "FOR UPDATE").
			Select("seq").
			Where("object_type = ? and object_id = ?", objectType, objectId).First(&seq).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在就插入
			err = tx.Create(&Seq{
				ObjectType: objectType,
				ObjectID:   objectId,
				Seq:        1,
			}).Error
		} else {
			// 有就 +1
			err = tx.Model(&Seq{}).Where("object_type = ? and object_id = ?", objectType, objectId).
				Update("seq", seq+1).Error
		}
		return err
	})
	if err != nil {
		return 0, err
	}
	return seq + 1, nil
}
