package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
)

//GetFollowUserIDListRedis 获取用户的关注用户ID列表
func (r *RelationService) GetFollowUserIDListRedis(userID uint64, followUserIDList *[]uint64) error {
	keyEmpty := fmt.Sprintf(utils.FollowEmptyPattern, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyFollow := fmt.Sprintf(utils.FollowPattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyFollow, utils.FollowExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		userIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyFollow, 0, -1).Result()
		if err != nil {
			return err
		}
		*followUserIDList = make([]uint64, 0, len(userIDStrList))
		for _, userIDStr := range userIDStrList {
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				continue
			}
			*followUserIDList = append(*followUserIDList, userID)
		}
		return nil
	}
	//缓存不存在，查询数据库
	var followUserList []system.Relation
	if err := global.DouYinDB.Where("from_user_id = ? and is_follow = ?", userID, true).Find(&followUserList).Error; err != nil {
		return err
	}
	if len(followUserList) == 0 { //关注数为0，做空值缓存处理,直接返回
		if err := r.SetUserFollowEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	listZ := make([]*redis.Z, 0, len(followUserList))
	*followUserIDList = make([]uint64, 0, len(followUserList))

	for _, user := range followUserList {
		*followUserIDList = append(*followUserIDList, user.ToUserID)
		listZ = append(listZ, &redis.Z{
			Score:  float64(user.UpdatedAt.UnixMilli() / 1000),
			Member: user.ToUserID,
		})
	}
	//将用户关注的用户ID列表加入缓存
	if err := r.SetFollowUserIDListRedis(userID, listZ...); err != nil {
		return err
	}
	return nil
}

//GetFollowUserListRedis 获取用户的关注列表
func (r *RelationService) GetFollowUserListRedis(userID uint64, followUserList *[]system.User) error {
	keyEmpty := fmt.Sprintf(utils.FollowEmptyPattern, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyFollow := fmt.Sprintf(utils.FollowPattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyFollow, utils.FollowExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		followUserIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyFollow, 0, -1).Result()
		if err != nil {
			return err
		}
		followUserIDList := make([]uint64, 0, len(followUserIDStrList))
		for _, userIDStr := range followUserIDStrList {
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				continue
			}
			followUserIDList = append(followUserIDList, userID)
		}
		if err := ServiceGroupApp.UserService.GetUserListByIDListRedis(followUserList, followUserIDList); err != nil {
			return err
		}
		return nil
	}
	//缓存不存在，查询数据库
	if err := r.GetFollowUserListSql(userID, followUserList); err != nil {
		return err
	}
	if len(*followUserList) == 0 { //用户关注数量为空，做空值缓存处理
		if err := r.SetUserFollowEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	return nil
}

//SetFollowUserIDListRedis 将用户关注的用户ID列表加入缓存
func (r *RelationService) SetFollowUserIDListRedis(userID uint64, listZ ...*redis.Z) error {
	//定义key
	keyFollow := fmt.Sprintf(utils.FollowPattern, userID)
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(global.DouYinCONTEXT, keyFollow, listZ...)
		pipe.Expire(global.DouYinCONTEXT, keyFollow, utils.FollowExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//SetUserFollowEmpty 设置用户关注列表为空的空值缓存处理
func (r *RelationService) SetUserFollowEmpty(userID uint64) error {
	keyEmpty := fmt.Sprintf(utils.FollowEmptyPattern, userID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.FollowEmptyExpire).Err()
}

//GetFollowerUserListRedis 获取用户的粉丝列表
func (r *RelationService) GetFollowerUserListRedis(userID uint64, followerUserList *[]system.User) error {
	keyEmpty := fmt.Sprintf(utils.FollowerEmptyPatten, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyFollower := fmt.Sprintf(utils.FollowerPattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyFollower, utils.FollowerExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		followerUserIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyFollower, 0, -1).Result()
		if err != nil {
			return err
		}
		followerUserIDList := make([]uint64, 0, len(followerUserIDStrList))
		for _, userIDStr := range followerUserIDStrList {
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				continue
			}
			followerUserIDList = append(followerUserIDList, userID)
		}
		if err := ServiceGroupApp.UserService.GetUserListByIDListRedis(followerUserList, followerUserIDList); err != nil {
			return err
		}
		return nil
	}
	//缓存不存在，查询数据库
	if err := r.GetFollowerUserListSql(userID, followerUserList); err != nil {
		return err
	}
	if len(*followerUserList) == 0 { //用户关注数量为空，做空值缓存处理
		if err := r.SetUserFollowerEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	return nil
}

//SetFollowerUserIDListRedis 将用户的粉丝ID列表加入缓存
func (r *RelationService) SetFollowerUserIDListRedis(userID uint64, listZ ...*redis.Z) error {
	//定义key
	keyFollower := fmt.Sprintf(utils.FollowerPattern, userID)
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(global.DouYinCONTEXT, keyFollower, listZ...)
		pipe.Expire(global.DouYinCONTEXT, keyFollower, utils.FollowerExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//SetUserFollowerEmpty 设置用户粉丝列表为空的空值缓存处理
func (r *RelationService) SetUserFollowerEmpty(userID uint64) error {
	keyEmpty := fmt.Sprintf(utils.FollowerEmptyPatten, userID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.FollowerEmptyExpire).Err()
}

//GetUserFollowStatusByUserIDRedis 获取用户对另一用户的关注状态
func (r *RelationService) GetUserFollowStatusByUserIDRedis(fromUserID, toUserID uint64) (bool, error) {
	//定义key
	keyFollow := fmt.Sprintf(utils.FollowPattern, fromUserID)
	lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])<=0 then
			return false
		end
		if redis.call("ZScore",KEYS[1],ARGV[2])==nil then
			return {err = "not favorite"}
		else 
			return true
		end`)
	keys := []string{keyFollow}
	values := []interface{}{utils.FollowExpire + utils.GetRandExpireTime(), toUserID}
	err := lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values).Err()
	if err == redis.Nil {
		return false, err
	} else if errors.Is(err, errors.New("not favorite")) {
		return false, nil
	} else {
		return true, nil
	}
}

//AddUserFollowRedis 关注成功后，如果缓存存在,更新用户关注数,用户关注列表
func (r *RelationService) AddUserFollowRedis(fromUserID, toUserID uint64) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		//定义key
		keyFollow := fmt.Sprintf(utils.FollowPattern, fromUserID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("HIncrBy",KEYS[1],"follow_count",1)
			return true
		end
		return false
		`)
		keys := []string{keyFollow}
		values := []interface{}{utils.FollowExpire + utils.GetRandExpireTime()}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	go func() {
		defer wg.Done()
		//定义key
		keyFollow := fmt.Sprintf(utils.FollowPattern, fromUserID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("ZAdd",KEYS[1],ARGV[2],ARGV[3])
			return true
		end
		return false
		`)
		keys := []string{keyFollow}
		values := []interface{}{utils.FollowExpire + utils.GetRandExpireTime(), float64(time.Now().UnixMilli() / 1000), toUserID}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	wg.Wait()
}

//SubUserFollowRedis 取消关注成功后，如果缓存存在,更新用户关注数,用户关注列表
func (r *RelationService) SubUserFollowRedis(fromUserID, toUserID uint64) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		//定义key
		keyFollow := fmt.Sprintf(utils.FollowPattern, fromUserID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("HIncrBy",KEYS[1],"follow_count",-1)
			return true
		end
		return false
		`)
		keys := []string{keyFollow}
		values := []interface{}{utils.FollowExpire + utils.GetRandExpireTime()}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	go func() {
		defer wg.Done()
		//定义key
		keyFollow := fmt.Sprintf(utils.FollowPattern, fromUserID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("ZRem",KEYS[1],ARGV[2])
			return true
		end
		return false
		`)
		keys := []string{keyFollow}
		values := []interface{}{utils.FollowExpire + utils.GetRandExpireTime(), toUserID}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	wg.Wait()
}
