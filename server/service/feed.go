package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type FeedService struct {
}

//GetStartTime 获取视频流Feed时间戳参数处理
func (f *FeedService) GetStartTime(startTime string) string {
	var latestTime string
	if startTime == "" || startTime == "0" {
		latestTime = strconv.FormatInt(time.Now().UnixMilli()/1000, 10) // 不传默认为当前服务器时间
	} else {
		lt, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			startTime = strconv.FormatInt(time.Now().UnixMilli()/1000, 10) // 解析出错默认为当前服务器时间
			return startTime
		}
		latestTime = strconv.FormatInt(lt, 10)
	}
	return latestTime
}

//GetVideoListByVideoIDListSql 当缓存未命中时，根据视频ID列表从数据库查询数据，并写入缓存
func (f *FeedService) GetVideoListByVideoIDListSql(videoList *[]system.Video, videoIDList []uint64) error {
	*videoList = make([]system.Video, 0, len(videoIDList))
	var getVideoList []system.Video
	if err := global.DouYinDB.Where("id in ?", videoIDList).Find(&getVideoList).Error; err != nil {
		return err
	}
	//对查询结果建立map映射关系
	mapVideoIDToVideo := make(map[uint64]system.Video, len(getVideoList))
	for i, video := range getVideoList {
		var status = true
		//查询favorite_count
		if err := global.DouYinDB.Model(&system.Favorite{}).Where("video_id = ? and is_favorite = ?", video.ID, &status).Count(&getVideoList[i].FavoriteCount).Error; err != nil {
			return err
		}
		//查询comment_count
		if err := global.DouYinDB.Model(&system.Comment{}).Where("video_id = ? ", video.ID).Count(&getVideoList[i].CommentCount).Error; err != nil {
			return err
		}
		mapVideoIDToVideo[video.ID] = getVideoList[i]
	}
	for _, videoID := range videoIDList {
		if v, ok := mapVideoIDToVideo[videoID]; ok {
			*videoList = append(*videoList, v)
		}
	}
	//将视频批量写入缓存
	return f.SetVideoListRedis(*videoList)
}

func (f *FeedService) GetAuthorIDByVideoID(videoID uint64) (authorID uint64, err error) {
	//查缓存
	keyVideo := fmt.Sprintf(utils.VideoPattern, videoID)
	result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime()).Result()
	if err == nil && result {
		s, err := global.DouYinRedis.HGet(global.DouYinCONTEXT, keyVideo, "author_id").Result()
		if err == nil {
			return strconv.ParseUint(s, 10, 64)
		}
	}
	//缓存不存在，查数据库
	var video system.Video
	if err = global.DouYinDB.Where("id = ?", videoID).First(&video).Error; err != nil {
		return 0, err
	}
	return video.AuthorID, nil
}
