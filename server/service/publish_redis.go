package service

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
)

//GetPublishedVideoListRedis 查询用户的发布视频列表
func (p *PublishService) GetPublishedVideoListRedis(userID uint64, videoList *[]system.Video) error {
	keyEmpty := fmt.Sprintf(utils.PublishEmptyPattern, userID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyPublish := fmt.Sprintf(utils.PublishPattern, userID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyPublish, utils.PublishExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		videoIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyPublish, 0, -1).Result()
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
	//缓存不存在，根据作者ID查询数据库
	if err = p.GetVideoListByUserIDSql(userID, videoList); err != nil {
		return err
	}
	if len(*videoList) == 0 { //用户发布视频数量为空，做空值缓存处理
		if err := p.SetUserPublishEmpty(userID); err != nil {
			return err
		}
		return nil
	}
	listZ := make([]*redis.Z, 0, len(*videoList))
	videoIDList := make([]uint64, 0, len(*videoList))
	for _, video := range *videoList {
		listZ = append(listZ, &redis.Z{
			Score:  float64(video.CreatedAt.UnixMilli() / 1000),
			Member: video.ID,
		})
		videoIDList = append(videoIDList, video.ID)
	}
	//将用户发表过的视频ID列表写入缓存
	if err := p.SetPublishedVideoIDListRedis(userID, listZ...); err != nil {
		return err
	}
	//将视频信息列表批量写入缓存
	if err := p.SetVideoListRedis(*videoList); err != nil {
		return err
	}
	return nil
}

//SetPublishedVideoIDListRedis 将用户发布过的视频ID列表加入缓存
func (p *PublishService) SetPublishedVideoIDListRedis(userID uint64, listZ ...*redis.Z) error {
	//定义key
	keyPublish := fmt.Sprintf(utils.PublishPattern, userID)
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(global.DouYinCONTEXT, keyPublish, listZ...)
		pipe.Expire(global.DouYinCONTEXT, keyPublish, utils.PublishExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//SetVideoListRedis 将视频信息列表批量写入缓存
func (p *PublishService) SetVideoListRedis(videoList []system.Video) error {
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		for _, video := range videoList {
			//定义video缓存中的key; 哈希结构 HMSet HGet HGetAll
			keyVideo := fmt.Sprintf(utils.VideoPattern, video.ID)
			pipe.HMSet(global.DouYinCONTEXT, keyVideo, "id", video.ID, "created_at", video.CreatedAt, "title", video.Title, "author_id", video.AuthorID, "play_url", video.PlayUrl, "cover_url", video.CoverUrl,
				"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount)
			pipe.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime())
		}
		return nil
	})
	return err
}

//SetUserPublishEmpty 设置用户发布视频列表为空的空值缓存处理
func (p *PublishService) SetUserPublishEmpty(userID uint64) error {
	keyEmpty := fmt.Sprintf(utils.PublishEmptyPattern, userID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.PublishEmptyExpire).Err()
}
