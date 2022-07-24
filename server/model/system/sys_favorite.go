package system

import (
	"time"

	"gorm.io/gorm"
)

//MySQl 建立的索引
//- id为主键，sonyflake雪花算法生成的分布式id
//- 联合索引 (user_id, video_id),(user_id,is_favorite)

//缓存层 Redis 中，
//- 有序集合 zset 记录用户历史点赞的所有视频（key: "favorite:"+user_id，score:点赞更新时间戳，value: video_id）。

//Favorite 点赞：数据库表favorites对应的model
type Favorite struct {
	ID         uint64         `gorm:"column:id"`          //点赞表ID
	CreatedAt  time.Time      `gorm:"column:created_at"`  //创建时间
	UpdatedAt  time.Time      `gorm:"column:updated_at"`  //更新时间
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at"`  //删除时间
	UserID     uint64         `gorm:"column:user_id"`     //点赞用户ID
	VideoID    uint64         `gorm:"column:video_id"`    //被点赞视频ID
	IsFavorite *bool          `gorm:"column:is_favorite"` //是否点赞,false-未点赞，true-点赞
}

func (f Favorite) TableName() string {
	return "favorites"
}
