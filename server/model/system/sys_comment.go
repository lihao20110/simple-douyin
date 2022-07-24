package system

import (
	"time"

	"gorm.io/gorm"
)

//MySQl 建立的索引
//- id为主键，sonyflake雪花算法生成的分布式id
//- 普通索引：video_id

//缓存层 Redis 中，
//- 使用有序集合 zset 来记录每个视频发布过的所有评论（key:"commentsOfVideo:"+video_id, score:评论创建时间戳， value:comment_id）。
//- 使用哈希结构 hset 记录每一个评论的相关信息（key:"comment:"+comment_id, field 包括 user_id, video_id, content, created_at等）。

//Comment 评论：数据库表comments对应的model
type Comment struct {
	ID        uint64         `gorm:"column:id" redis:"id"`             //评论ID
	CreatedAt time.Time      `gorm:"column:created_at" redis:"-"`      //创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at" redis:"-"`      //更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" redis:"-"`      //删除时间
	Content   string         `gorm:"column:content" redis:"content"`   //评论内容
	UserID    uint64         `gorm:"column:user_id" redis:"user_id"`   //评论所属视频ID
	VideoID   uint64         `gorm:"column:video_id" redis:"video_id"` //评论所属用户ID
}

func (c Comment) TableName() string {
	return "comments"
}
