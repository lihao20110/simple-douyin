package system

import (
	"time"

	"gorm.io/gorm"
)

// 数据库表users对应的model
type User struct {
	ID            uint64         `gorm:"column:user_id;primaryKey"` //用户ID
	CreatedAt     time.Time      `gorm:"column:created_at"`         //创建时间
	UpdatedAt     time.Time      `gorm:"column:updated_at"`         //更新时间
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`   //删除时间
	UserName      string         `gorm:"column:user_name"`          //用户名
	PassWord      string         `gorm:"column:user_password"`      //用户密码
	FollowCount   uint64         `gorm:"column:follow_count"`       //关注数量
	FollowerCount uint64         `gorm:"column:follower_count"`     //粉丝数量
}
