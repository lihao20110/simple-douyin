package service

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
)

//GetFeedVideoListRedis 拉取视频数据,返回按投稿时间倒序的视频列表
func (f *FeedService) GetFeedVideoListRedis(videoList *[]system.Video, authorList *[]system.User, latestTime string) error {
	//1.确保feed已缓存在redis中
	if err := f.SetFeedCache(); err != nil {
		return err
	}
	//2.接下来，从feed缓存中取出返回给用户的视频ID列表数据
	startTime := f.GetStartTime(latestTime) //时间戳字符串
	opt := redis.ZRangeBy{                  // 初始化查询条件， Offset和Count用于分页
		Min:    "0",                 //最小分数
		Max:    startTime,           // 最大分数
		Offset: 0,                   // 类似MySQL的limit, 表示开始偏移量
		Count:  global.FeedVideoNum, // 返回多少数据
	}
	// ZREVRANGE命令 成员按 score 值递减(从大到小)来排列，这里指视频发布日期
	videoIDStrList, err := global.DouYinRedis.ZRevRangeByScore(global.DouYinCONTEXT, "feed", &opt).Result()
	if err != nil {
		return err
	}
	if len(videoIDStrList) == 0 {
		return nil
	}
	videoIDList := make([]uint64, 0, len(videoIDStrList))
	for _, videoStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoStr, 10, 64)
		if err != nil {
			continue
		}
		videoIDList = append(videoIDList, videoID)
	}
	//3.根据以上取得的视频ID列表数据，查询视频详情数据
	if err := f.GetVideoListByVideoIDListRedis(videoList, videoIDList); err != nil {
		return err
	}
	//批量获取视频作者信息
	authorIDList := make([]uint64, len(videoIDList))
	for i, video := range *videoList {
		authorIDList[i] = video.AuthorID
	}
	if err := ServiceGroupApp.UserService.GetUserListByIDListRedis(authorList, authorIDList); err != nil {
		return err
	}
	return nil
}

//GetVideoListByVideoIDListRedis 根据视频ID列表从缓存中查找视频信息
func (f *FeedService) GetVideoListByVideoIDListRedis(videoList *[]system.Video, videoIDList []uint64) error {
	*videoList = make([]system.Video, 0, len(videoIDList))
	inCache := make([]bool, 0, len(videoIDList))
	notInCacheIDList := make([]uint64, 0, len(videoIDList))
	for _, videoID := range videoIDList {
		//定义key; 哈希结构 HMSet HGet HGetAll
		keyVideo := fmt.Sprintf(utils.VideoPattern, videoID)
		//1.先直接使用命令Expire判断并更新过期时间，不推荐使用Exists
		result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if !result { //当前视频不在缓存中
			*videoList = append(*videoList, system.Video{})
			inCache = append(inCache, false)
			notInCacheIDList = append(notInCacheIDList, videoID)
			continue
		}
		var video system.Video
		//2.取数据
		if err := global.DouYinRedis.HGetAll(global.DouYinCONTEXT, keyVideo).Scan(&video); err != nil {
			return err
		}
		//error-redis.Scan(unsupported time.Time)
		time, err := global.DouYinRedis.HGet(global.DouYinCONTEXT, keyVideo, "created_at").Time()
		if err != nil {
			continue
		}
		video.CreatedAt = time
		*videoList = append(*videoList, video)
		inCache = append(inCache, true)
	}
	if len(notInCacheIDList) == 0 {
		return nil //所需视频全部在缓存中，提前返回
	}
	//从MySQL数据库批量查询不在redis缓存中的视频数据
	var notInCacheVideoList []system.Video
	if err := f.GetVideoListByVideoIDListSql(&notInCacheVideoList, notInCacheIDList); err != nil {
		return err
	}
	//加上查询到的数据
	for i, j := 0, 0; i < len(*videoList); i++ {
		if inCache[i] == false {
			(*videoList)[i] = notInCacheVideoList[j]
			j++
		}
	}
	return nil
}

//SetVideoListRedis 将视频信息列表批量写入缓存
func (f *FeedService) SetVideoListRedis(videoList []system.Video) error {
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

//SetFeedCache 使用有序集合 zset 记录所有发布过的视频（key:"feed"，score:视频创建的时间戳，value:video_id）
func (f *FeedService) SetFeedCache() error {
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, utils.FeedPattern).Result()
	if err != nil {
		return err
	}
	//如果 key 不存在，则创建一个空的有序集并执行 ZADD 操作
	if result <= 0 { //feed不存在
		var allVideos []system.Video
		if err := global.DouYinDB.Find(&allVideos).Error; err != nil {
			return err
		}
		if len(allVideos) == 0 {
			return nil //数据库查询无数据直接返回
		}
		//将数据库中的视频ID全部加入到缓存中
		listZ := make([]*redis.Z, 0, len(allVideos))
		for _, video := range allVideos {
			listZ = append(listZ, &redis.Z{
				Score:  float64(video.CreatedAt.UnixMilli() / 1000),
				Member: video.ID,
			})
		}
		return global.DouYinRedis.ZAdd(global.DouYinCONTEXT, "feed", listZ...).Err()
	}
	return nil
}
