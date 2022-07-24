package system

import (
	"time"

	"gorm.io/gorm"
)

//MySQl 建立的索引
//- id为主键，sonyflake雪花算法生成的分布式id
//- 联合索引 (from_user_id, to_user_id)，(from_user_id, is_follow),(to_user_id,is_follow)

//缓存层 Redis 中，
//- 使用有序集合 zset 来记录用户历史关注列表（key: "follow:"+user_id，score: 关注更新时间戳,value: user_id）
//- 使用有序集合 zset 来记录用户历史粉丝列表（key: "follower:"+user_id，score: 关注更新时间戳，value: user_id）

//Relation 关注：数据库表relations对应的model
type Relation struct {
	ID         uint64         `gorm:"column:id"`           //关系表ID
	CreatedAt  time.Time      `gorm:"column:created_at"`   //创建时间
	UpdatedAt  time.Time      `gorm:"column:updated_at"`   //更新时间
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at"`   //删除时间
	FromUserID uint64         `gorm:"column:from_user_id"` //关注者ID
	ToUserID   uint64         `gorm:"column:to_user_id"`   //被关注者ID
	IsFollow   *bool          `gorm:"column:is_follow"`    //是否关注,false-未关注，true-关注
}

func (r Relation) TableName() string {
	return "relations"
}
