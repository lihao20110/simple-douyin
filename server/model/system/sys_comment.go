package system

import (
	"time"

	"gorm.io/gorm"
)

// 数据库表comments对应的model
type Comment struct {
	ID        uint64         `gorm:"column:comment_id;primaryKey"` //评论ID
	CreatedAt time.Time      `gorm:"column:created_at"`            //创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at"`            //更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`      //删除时间
	Content   string         `gorm:"column:content"`               //评论内容
	UserId    uint64         `gorm:"column:user_id"`               //评论所属视频ID
	VideoId   uint64         `gorm:"column:video_id"`              //评论所属用户ID
}
