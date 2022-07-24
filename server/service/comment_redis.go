package service

import "C"
import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
)

//SetVideoCommentEmpty 设置视频评论列表为空的空值缓存处理
func (f *FavoriteService) SetVideoCommentEmpty(videoID uint64) error {
	keyEmpty := fmt.Sprintf(utils.CommentEmptyPattern, videoID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.CommentEmptyExpire).Err()
}

//GetVideoCommentListRedis 获取视频的评论列表
func (c *CommentService) GetVideoCommentListRedis(videoID uint64, commentModelList *[]system.Comment) error {
	keyEmpty := fmt.Sprintf(utils.CommentEmptyPattern, videoID)
	result, err := global.DouYinRedis.Exists(global.DouYinCONTEXT, keyEmpty).Result()
	if err != nil {
		return err
	}
	if result > 0 { //存在空值缓存，直接返回
		return nil
	}
	keyComments := fmt.Sprintf(utils.CommentsOfVideoPattern, videoID)
	ok, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyComments, utils.CommentsOfVideoExpire+utils.GetRandExpireTime()).Result()
	if err != nil {
		return err
	}
	if ok { //缓存存在，直接返回
		commentIDStrList, err := global.DouYinRedis.ZRevRange(global.DouYinCONTEXT, keyComments, 0, -1).Result()
		if err != nil {
			return err
		}
		commentIDList := make([]uint64, 0, len(commentIDStrList))
		for _, videoIDStr := range commentIDStrList {
			videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
			if err != nil {
				continue
			}
			commentIDList = append(commentIDList, videoID)
		}
		if err = c.GetCommentListByCommentIDListRedis(commentModelList, commentIDList); err != nil {
			return err
		}
		return nil
	}
	//缓存不存在，查询数据库
	if err = c.GetVideoCommentListByVideoIDSql(videoID, commentModelList); err != nil {
		return err
	}
	if len(*commentModelList) == 0 { //视频评论数量为空，做空值缓存处理
		if err := c.SetVideoCommentEmpty(videoID); err != nil {
			return err
		}
		return nil
	}
	return nil
}

//SetVideoCommentIDListRedis 将视频的评论ID列表加入缓存
func (c *CommentService) SetVideoCommentIDListRedis(videoID uint64, listZ ...*redis.Z) error {
	//定义key
	keyFavorite := fmt.Sprintf(utils.FavoritePattern, videoID)
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(global.DouYinCONTEXT, keyFavorite, listZ...)
		pipe.Expire(global.DouYinCONTEXT, keyFavorite, utils.FavoriteExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//SetVideoCommentEmpty 设置视频评论列表为空的空值缓存处理
func (c *CommentService) SetVideoCommentEmpty(videoID uint64) error {
	keyEmpty := fmt.Sprintf(utils.CommentEmptyPattern, videoID)
	return global.DouYinRedis.Set(global.DouYinCONTEXT, keyEmpty, "1", utils.CommentEmptyExpire).Err()
}

//SetCommentListRedis 将评论信息列表批量写入缓存
func (c *CommentService) SetCommentListRedis(commentList []system.Comment) error {
	//TxPipeline
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		for _, comment := range commentList {
			//定义comment缓存中的key; 哈希结构 HMSet HGet HGetAll
			keyComment := fmt.Sprintf(utils.CommentPattern, comment.ID)
			pipe.HMSet(global.DouYinCONTEXT, keyComment, "id", comment.ID, "created_at", comment.CreatedAt.String(), "content", comment.Content, "user_id", comment.UserID, "video_id", comment.VideoID)
			pipe.Expire(global.DouYinCONTEXT, keyComment, utils.CommentExpire+utils.GetRandExpireTime())
		}
		return nil
	})
	return err
}

//GetCommentListByCommentIDListRedis 根据评论ID列表从缓存中查找评论信息
func (c *CommentService) GetCommentListByCommentIDListRedis(commentModelList *[]system.Comment, commentIDList []uint64) error {
	*commentModelList = make([]system.Comment, 0, len(commentIDList))
	inCache := make([]bool, 0, len(commentIDList))
	notInCacheIDList := make([]uint64, 0, len(commentIDList))
	for _, commentID := range commentIDList {
		//定义key; 哈希结构 HMSet HGet HGetAll
		keyComment := fmt.Sprintf(utils.CommentPattern, commentID)
		//1.先直接使用命令Expire判断并更新过期时间，不推荐使用Exists
		result, err := global.DouYinRedis.Expire(global.DouYinCONTEXT, keyComment, utils.CommentExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if !result { //当前评论不在缓存中
			*commentModelList = append(*commentModelList, system.Comment{})
			inCache = append(inCache, false)
			notInCacheIDList = append(notInCacheIDList, commentID)
			continue
		}
		var comment system.Comment
		//2.取数据
		if err := global.DouYinRedis.HGetAll(global.DouYinCONTEXT, keyComment).Scan(&comment); err != nil {
			return err
		}
		//error-redis.Scan(unsupported time.Time)
		time, err := global.DouYinRedis.HGet(global.DouYinCONTEXT, keyComment, "created_at").Time()
		if err != nil {
			continue
		}
		comment.CreatedAt = time
		*commentModelList = append(*commentModelList, comment)
		inCache = append(inCache, true)
	}
	if len(notInCacheIDList) == 0 {
		return nil //所需视频全部在缓存中，提前返回
	}
	//从MySQL数据库批量查询不在redis缓存中的视频数据
	var notInCacheCommentList []system.Comment
	if err := c.GetCommentListByCommentIDListSql(&notInCacheCommentList, notInCacheIDList); err != nil {
		return err
	}
	//加上查询到的数据
	j := 0
	for i := 0; i < len(*commentModelList); i++ {
		if inCache[i] == false {
			(*commentModelList)[i] = notInCacheCommentList[j]
			j++
		}
	}
	return nil
}

//DeleteCommentRedis 删除评论的缓存操作
func (c *CommentService) DeleteCommentRedis(commentID, videoID uint64) error {
	//定义key
	keyCommentsOfVideo := fmt.Sprintf(utils.CommentsOfVideoPattern, videoID)
	keyVideo := fmt.Sprintf(utils.VideoPattern, videoID)
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		result, err := pipe.Expire(global.DouYinCONTEXT, keyCommentsOfVideo, utils.CommentsOfVideoExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if result {
			_, err := pipe.ZRem(global.DouYinCONTEXT, keyCommentsOfVideo, commentID).Result()
			if err != nil {
				return err
			}
		}
		result, err = pipe.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if result {
			_, err := pipe.HIncrBy(global.DouYinCONTEXT, keyVideo, "comment_count", -1).Result()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

//AddCommentRedis 添加评论的缓存操作
func (c *CommentService) AddCommentRedis(commentModel *system.Comment) error {
	//定义key
	keyEmpty := fmt.Sprintf(utils.CommentEmptyPattern, commentModel.VideoID)
	if err := global.DouYinRedis.Del(global.DouYinCONTEXT, keyEmpty).Err(); err != nil {
		return err
	}
	keyCommentsOfVideo := fmt.Sprintf(utils.CommentsOfVideoPattern, commentModel.VideoID)
	keyVideo := fmt.Sprintf(utils.VideoPattern, commentModel.VideoID)
	keyComment := fmt.Sprintf(utils.CommentPattern, commentModel.ID)
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		result, err := pipe.Expire(global.DouYinCONTEXT, keyCommentsOfVideo, utils.CommentsOfVideoExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if result {
			_, err := pipe.ZAdd(global.DouYinCONTEXT, keyCommentsOfVideo, &redis.Z{
				Score:  float64(commentModel.CreatedAt.UnixMilli() / 1000),
				Member: commentModel.ID,
			}).Result()
			if err != nil {
				return err
			}
		}
		result, err = pipe.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		if result {
			_, err := pipe.HIncrBy(global.DouYinCONTEXT, keyVideo, "comment_count", 1).Result()
			if err != nil {
				return err
			}
		}
		_, err = pipe.HMSet(global.DouYinCONTEXT, keyComment, "id", commentModel.ID, "user_id", commentModel.UserID, "video_id", commentModel.VideoID, "content", commentModel.Content, "created_at", commentModel.CreatedAt.UnixMilli()).Result()
		if err != nil {
			return err
		}
		_, err = pipe.Expire(global.DouYinCONTEXT, keyComment, utils.CommentExpire+utils.GetRandExpireTime()).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
