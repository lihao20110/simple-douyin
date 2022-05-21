package global

import (
	"github.com/go-redis/redis/v8"
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
	DouYinConcurrencyControl = &singleflight.Group{}
)

const (
	ConfigEnv      = "DouYinCONFIG"
	ConfigFile     = "config.yaml"
	FeedVideoNum   = 30                    //刷视频，每次下拉刷新视频数目
	MaxUserLength  = 30                    //用户名最大长度
	MinUserLength  = 3                     //用户名最小长度
	PassWordRegexp = "^[a-zA-Z]\\w{5,17}$" //以字母开头，长度在6~18之间，只能包含字符、数字和下划线。
	MaxTitleLength = 30                    //视频标题描述最大字数
	MinTitleLength = 3                     //视频标题描述最小字数
	MB             = 1024 * 1024
)
