package utils

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
)

//在redis中缓存token，并设置过期时间
func SetToken(token string, userId uint64) error {
	err := global.DouYinRedis.Set(context.Background(), token, userId, time.Duration(global.DouYinCONFIG.JWT.ExpiresTime)*time.Second).Err() // 此处过期时间等于jwt过期时间
	return err
}

//通过token在redis查询用户id
func GetUserId(token string) uint64 {
	userIdStr, err := global.DouYinRedis.Get(context.Background(), token).Result()
	if err == redis.Nil {
		//global.DouYinLOG.Info(err.Error())
		return 0
	} else if err != nil {
		panic(err)
	}
	//global.DouYinLOG.Info(userIdStr, zap.String(userIdStr, userIdStr))
	userId, _ := strconv.ParseUint(userIdStr, 10, 64)
	//global.DouYinLOG.Info(string(userId), zap.String(string(userId), string(userId)))
	return userId
}
