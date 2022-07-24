package system

import (
	"time"

	"gorm.io/gorm"
)

//MySQl 建立的索引
//- id为主键，sonyflake雪花算法生成的分布式id
//- 唯一索引：用户名name

//缓存层 Redis 中，
//- 使用哈希结构 hset 记录用户的相关信息，（key:"user:"+user_id，field 包括 follow_count, follower_count等）。

//User 用户：数据库表users对应的model
type User struct {
	ID            uint64         `gorm:"column:id" redis:"id"`             //用户ID
	CreatedAt     time.Time      `gorm:"column:created_at" redis:"-"`      //创建时间
	UpdatedAt     time.Time      `gorm:"column:updated_at" redis:"-"`      //更新时间
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at" redis:"-"`      //删除时间
	Name          string         `gorm:"column:name" redis:"name"`         //用户名
	PassWord      string         `gorm:"column:password" redis:"password"` //用户密码
	FollowCount   int64          `gorm:"-" redis:"follow_count"`           //关注数量
	FollowerCount int64          `gorm:"-" redis:"follower_count"`         //粉丝数量
}

func (u User) TableName() string {
	return "users"
}
