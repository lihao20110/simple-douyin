package global

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sony/sonyflake"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/lihao20110/simple-douyin/server/config"
)

var (
	DouYinDB                 *gorm.DB
	DouYinRedis              *redis.Client
	DouYinCONFIG             config.Server
	DouYinVP                 *viper.Viper
	DouYinLOG                *zap.Logger
	DouYinIDGenerator        *sonyflake.Sonyflake
	DouYinCONTEXT            = context.Background() // 上下文信息
	DouYinConcurrencyControl = &singleflight.Group{}
)

const (
	ConfigEnv        = "DouYinCONFIG"
	ConfigFile       = "config.yaml"
	FeedVideoNum     = 30                    //刷视频，每次下拉刷新视频数目
	MaxUserLength    = 30                    //用户名最大长度
	MinUserLength    = 3                     //用户名最小长度
	PassWordRegexp   = "^[a-zA-Z]\\w{5,17}$" //以字母开头，长度在6~18之间，只能包含字符、数字和下划线。
	MaxTitleLength   = 30                    //视频标题描述最大字数
	MinTitleLength   = 3                     //视频标题描述最小字数
	MB               = 1024 * 1024           //1MB = 1024*1024B
	StartTime        = "2022-05-01 00:00:01" // 固定启动时间，保证生成 ID 唯一性
	MaxCommentLength = 191                   //评论描述最大字数
	MinCommentLength = 1                     //评论描述最小字数
)
