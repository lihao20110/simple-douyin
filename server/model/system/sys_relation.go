package system

import (
	"time"

	"gorm.io/gorm"
)

// 数据库表relations对应的model
type Relation struct {
	ID         uint64         `gorm:"column:relation_id;primaryKey"` //关系表ID
	CreatedAt  time.Time      `gorm:"column:created_at"`             //创建时间
	UpdatedAt  time.Time      `gorm:"column:updated_at"`             //更新时间
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`       //删除时间
	FromUserId uint64         `gorm:"column:from_user_id"`           //关注者ID
	ToUserId   uint64         `gorm:"column:to_user_id"`             //被关注者ID
	IsMutual   uint64         `gorm:"column:is_mutual"`              //是否关注,0-未关注，1-关注
}
