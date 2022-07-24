package service

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"gorm.io/gorm"
)

//GetUserListByIDListRedis 根据用户ID列表查找一批用户信息
func (u *UserService) GetUserListByIDListRedis(authorList *[]system.User, authorIDList []uint64) error {
	//1.先查缓存，未查到(部分)则从数据库中查，并写入缓存
	*authorList = make([]system.User, 0, len(authorIDList))
	inCache := make([]bool, 0, len(authorIDList))
	notInCacheIDList := make([]uint64, 0, len(authorIDList))
	for _, userID := range authorIDList {
		//定义key; 哈希结构 HMSet HGet HGetAll
		keyUser := fmt.Sprintf(utils.UserPattern, userID)
		//1.先直接使用命令Expire判断并更新过期时间，不推荐使用Exists
		result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyUser, utils.UserExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if !result { //当前用户信息不在缓存中
			*authorList = append(*authorList, system.User{})
			inCache = append(inCache, false)
			notInCacheIDList = append(notInCacheIDList, userID)
			continue
		}
		var user system.User
		//2.取数据
		if err := global.DouYinRedis.HGetAll(global.DouYinCONTEXT, keyUser).Scan(&user); err != nil {
			return err
		}
		*authorList = append(*authorList, user)
		inCache = append(inCache, true)
	}
	if len(notInCacheIDList) == 0 {
		return nil //所需用户信息全部在缓存中，提前返回
	}
	//2.从MySQL数据库批量查询不在redis缓存中的用户数据
	var notInCacheUserList []system.User
	if err := u.GetUserListByUserIDListSql(&notInCacheUserList, notInCacheIDList); err != nil {
		return err
	}
	//加上查询到的数据
	for i, j := 0, 0; i < len(*authorList); i++ {
		if inCache[i] == false {
			(*authorList)[i] = notInCacheUserList[j]
			j++
		}
	}
	return nil
}

//GetUserInfoByUserIDRedis 查询单个用户详细信息
func (u *UserService) GetUserInfoByUserIDRedis(userID uint64, userInfo *system.User) error {
	//先查缓存
	keyUserInfoEmpty := fmt.Sprintf(utils.UserInfoEmptyPattern, userID)
	if result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyUserInfoEmpty, utils.UserInfoEmptyExpire).Result(); err == nil && result {
		return errors.New("userInfo not exists")
	}
	keyUser := fmt.Sprintf(utils.UserPattern, userID) //根据规则定义key
	//1.先直接使用命令Expire判断并更新过期时间，不推荐使用Exists
	result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyUser, utils.UserExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if result { //在缓存中,取数据
		if err := global.DouYinRedis.HGetAll(global.DouYinCONTEXT, keyUser).Scan(userInfo); err != nil {
			return err
		}
		return nil
	}
	//2.没有命中缓存，从MySQL数据库中查询
	if err := u.GetUserInfoByUserIDListSql(userID, userInfo); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { //缓存穿透，空值处理
			global.DouYinRedis.Set(global.DouYinCONTEXT, keyUserInfoEmpty, "1", utils.UserInfoEmptyExpire)
			return errors.New("userInfo not exists")
		}
		return err
	}
	//3.将数据写入缓存
	if err := u.SetUserInfoToRedis(userInfo); err != nil {
		return err
	}
	return nil
}

//SetUserInfoToRedis //将userInfo数据写入缓存
func (u *UserService) SetUserInfoToRedis(user *system.User) error {
	keyUser := fmt.Sprintf(utils.UserPattern, user.ID)
	// 使用TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.HMSet(global.DouYinCONTEXT, keyUser, "id", user.ID, "name", user.Name, "password", user.PassWord, "follow_count", user.FollowCount, "follower_count", user.FollowerCount)
		pipe.Expire(global.DouYinCONTEXT, keyUser, utils.UserExpire+utils.GetRandExpireTime()) //设置过期时间
		return nil
	})
	return err
}

//SetUserListToRedis 将用户信息批量写入缓存
func (u *UserService) SetUserListToRedis(userList *[]system.User) error {
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		for _, user := range *userList {
			//定义User缓存中的key; 哈希结构 HMSet HGet HGetAll
			keyUser := fmt.Sprintf(utils.UserPattern, user.ID)
			pipe.HMSet(global.DouYinCONTEXT, keyUser, "id", user.ID, "name", user.Name, "password", user.PassWord, "follow_count", user.FollowCount, "follower_count", user.FollowerCount)
			pipe.Expire(global.DouYinCONTEXT, keyUser, utils.UserExpire+utils.GetRandExpireTime())
		}
		return nil
	})
	return err
}
