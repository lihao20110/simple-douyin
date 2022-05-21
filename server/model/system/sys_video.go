package system

import (
	"time"

	"gorm.io/gorm"
)

// 数据库表videos对应的model
type Video struct {
	ID            uint64         `gorm:"column:video_id;primaryKey"` //视频ID
	CreatedAt     time.Time      `gorm:"column:created_at"`          //视频发布创建时间
	UpdatedAt     time.Time      `gorm:"column:updated_at"`          //更新时间
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`    //删除时间
	Title         string         `gorm:"column:video_title" `        //视频标题
	UserId        uint64         `gorm:"column:author_id"`           //视频发布者ID
	PlayUrl       string         `gorm:"column:play_url"`            //视频源地址
	CoverUrl      string         `gorm:"column:cover_url"`           //视频封面地址
	FavoriteCount uint64         `gorm:"column:favorite_count"`      //视频点赞数
	CommentCount  uint64         `gorm:"column:comment_count"`       //视频评论数
}
