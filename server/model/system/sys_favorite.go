package system

import (
	"time"

	"gorm.io/gorm"
)

// 数据库表favorites对应的model
type Favorite struct {
	ID        uint64         `gorm:"column:favorite_id;primaryKey"` //点赞表ID
	CreatedAt time.Time      `gorm:"column:created_at"`             //创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at"`             //更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`       //删除时间
	Status    uint64         `gorm:"column:status"`                 //是否点赞,0-未点赞，1-点赞
	UserId    uint64         `gorm:"column:user_id"`                //点赞用户ID
	VideoId   uint64         `gorm:"column:video_id"`               //被点赞视频ID
}
