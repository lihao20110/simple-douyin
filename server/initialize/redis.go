package initialize

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"go.uber.org/zap"
)

func Redis() *redis.Client {
	redisConf := global.DouYinCONFIG.Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConf.Addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	})
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		global.DouYinLOG.Error("redis connect ping failed,err:", zap.Error(err))
		return nil
	}
	global.DouYinLOG.Info("redis connect ping success:", zap.String("pong", pong))
	return rdb
}
