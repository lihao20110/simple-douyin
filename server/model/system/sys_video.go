package system

import (
	"time"

	"gorm.io/gorm"
)

//MySQl 建立的索引
//- id为主键，sonyflake雪花算法生成的分布式id
//- 普通索引：发布者author_id

//缓存层redis中
//- 使用有序集合zset记录所有发布过的视频（key:"feed"，score:视频创建时间戳，value:video_id）。
//- 使用哈希结构hset记录每一个视频的相关信息（key:"video:"+video_id, field包括favorite_count, comment_count等）。
//- 使用有序集合zset记录每个用户发布过的所有视频（key:"publish:"+user_id, score:视频创建时间戳， value:video_id）。

//Video 视频：数据库表videos对应的model
type Video struct {
	ID            uint64         `gorm:"column:id" redis:"id"`               //视频ID
	CreatedAt     time.Time      `gorm:"column:created_at" redis:"-"`        //视频发布创建时间
	UpdatedAt     time.Time      `gorm:"column:updated_at" redis:"-"`        //更新时间
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at" redis:"-"`        //删除时间
	Title         string         `gorm:"column:title" redis:"title"`         //视频标题
	AuthorID      uint64         `gorm:"column:author_id" redis:"author_id"` //视频发布者ID
	PlayUrl       string         `gorm:"column:play_url" redis:"play_url"`   //视频源地址
	CoverUrl      string         `gorm:"column:cover_url" redis:"cover_url"` //视频封面地址
	FavoriteCount int64          `gorm:"-" redis:"favorite_count"`           //视频点赞数
	CommentCount  int64          `gorm:"-" redis:"comment_count" `           //视频评论数
}

func (v Video) TableName() string {
	return "videos"
}
