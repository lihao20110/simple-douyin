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

//GetFavoriteVideoIDListRedis 获取用户的点赞视频ID列表
func (f *FavoriteService) GetFavoriteVideoIDListRedis(userID uint64, videoIDList *[]uint64) error {
	keyEmpty := fmt.Sprintf(utils.FavoriteEmptyPattern, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyFavorite, utils.FavoriteExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		videoIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyFavorite, 0, -1).Result()
		if err != nil {
			return err
		}
		*videoIDList = make([]uint64, 0, len(videoIDStrList))
		for _, videoIDStr := range videoIDStrList {
			videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
			if err != nil {
				continue
			}
			*videoIDList = append(*videoIDList, videoID)
		}
		return nil
	}
	//缓存不存在，查询数据库
	var favoriteVideoList []system.Favorite
	if err := global.DouYinDB.Where("user_id = ? and is_favorite = ?", userID, true).Find(&favoriteVideoList).Error; err != nil {
		return err
	}
	if len(favoriteVideoList) == 0 { //点赞视频数为0，做空值缓存处理,直接返回
		if err := f.SetUserFavoriteEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	listZ := make([]*redis.Z, 0, len(favoriteVideoList))
	*videoIDList = make([]uint64, 0, len(favoriteVideoList))

	for _, video := range favoriteVideoList {
		*videoIDList = append(*videoIDList, video.ID)
		listZ = append(listZ, &redis.Z{
			Score:  float64(video.UpdatedAt.Unix()),
			Member: video.VideoID,
		})
	}
	//将用户点赞过的视频ID列表加入缓存
	if err := f.SetFavoriteVideoIDListRedis(userID, listZ...); err != nil {
		return err
	}
	return nil
}

//GetFavoriteVideoListRedis 获取用户的点赞视频列表
func (f *FavoriteService) GetFavoriteVideoListRedis(userID uint64, videoList *[]system.Video) error {
	keyEmpty := fmt.Sprintf(utils.FavoriteEmptyPattern, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyFavorite, utils.FavoriteExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		videoIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyFavorite, 0, -1).Result()
		if err != nil {
			return err
		}
		videoIDList := make([]uint64, 0, len(videoIDStrList))
		for _, videoIDStr := range videoIDStrList {
			videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
			if err != nil {
				continue
			}
			videoIDList = append(videoIDList, videoID)
		}
		if err = ServiceGroupApp.FeedService.GetVideoListByVideoIDListRedis(videoList, videoIDList); err != nil {
			return err
		}
		return nil
	}
	//缓存不存在，查询数据库
	if err = f.GetFavoriteVideoListByUserIDSql(userID, videoList); err != nil {
		return err
	}
	if len(*videoList) == 0 { //用户点赞视频数量为空，做空值缓存处理
		if err := f.SetUserFavoriteEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	return nil
}

//SetFavoriteVideoIDListRedis 将用户点赞过的视频ID列表加入缓存
func (f *FavoriteService) SetFavoriteVideoIDListRedis(userID uint64, listZ ...*redis.Z) error {
	//定义key
	keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(global.DouYinCONTEXT, keyFavorite, listZ...)
		pipe.Expire(global.DouYinCONTEXT, keyFavorite, utils.FavoriteExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//SetUserFavoriteEmpty 设置用户点赞视频列表为空的空值缓存处理
func (f *FavoriteService) SetUserFavoriteEmpty(userID uint64) error {
	keyEmpty := fmt.Sprintf(utils.FavoriteEmptyPattern, userID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.FavoriteEmptyExpire).Err()
}

//AddVideoFavoriteRedis 点赞成功后，如果缓存存在,更新视频被点赞数,用户点赞列表
func (f *FavoriteService) AddVideoFavoriteRedis(userID, videoID uint64) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		//定义key
		keyVideo := fmt.Sprintf(utils.VideoPattern, videoID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("HIncrBy",KEYS[1],"favorite_count",1)
			return true
		end
		return false
		`)
		keys := []string{keyVideo}
		values := []interface{}{utils.VideoExpire + utils.GetRandExpireTime()}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	go func() {
		defer wg.Done()
		//定义key
		keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("ZAdd",KEYS[1],ARGV[2],ARGV[3])
			return true
		end
		return false
		`)
		keys := []string{keyFavorite}
		values := []interface{}{utils.FavoriteExpire + utils.GetRandExpireTime(), float64(time.Now().Unix()), videoID}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	wg.Wait()
}

//SubVideoFavoriteRedis 取消点赞成功后，如果缓存存在,更新视频被点赞数,用户点赞列表
func (f *FavoriteService) SubVideoFavoriteRedis(userID, videoID uint64) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		//定义key
		keyVideo := fmt.Sprintf(utils.VideoPattern, videoID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("HIncrBy",KEYS[1],"favorite_count",-1)
			return true
		end
		return false
		`)
		keys := []string{keyVideo}
		values := []interface{}{utils.VideoExpire + utils.GetRandExpireTime()}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	go func() {
		defer wg.Done()
		//定义key
		keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
		lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])>0 then
			redis.call("ZRem",KEYS[1],ARGV[2])
			return true
		end
		return false
		`)
		keys := []string{keyFavorite}
		values := []interface{}{utils.FavoriteExpire + utils.GetRandExpireTime(), videoID}
		lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values)
	}()
	wg.Wait()
}

//GetVideoFavoriteStatusByUserIDRedis 获取用户对视频的点赞状态
func (f *FavoriteService) GetVideoFavoriteStatusByUserIDRedis(userID, videoID uint64) (bool, error) {
	//定义key
	keyFavorite := fmt.Sprintf(utils.FavoritePattern, userID)
	lua := redis.NewScript(`
		if redis.call("Expire",KEYS[1],ARGV[1])<=0 then
			return false
		end
		if redis.call("ZScore",KEYS[1],ARGV[2])==nil then
			return {err = "not favorite"}
		else 
			return true
		end`)
	keys := []string{keyFavorite}
	values := []interface{}{utils.FavoriteExpire + utils.GetRandExpireTime(), videoID}
	err := lua.Run(global.DouYinCONTEXT, global.DouYinRedis, keys, values).Err()
	if err == redis.Nil {
		return false, err
	} else if errors.Is(err, errors.New("not favorite")) {
		return false, nil
	} else {
		return true, nil
	}
}
