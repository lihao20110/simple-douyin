package service

import (
	"errors"
	"sort"
	"sync"

	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FavoriteService struct {
}

//点赞列表
func (f *FavoriteService) FavoriteList(userId uint64) (*[]comRes.Video, error) {
	var videoIds []uint
	err := global.DouYinDB.Model(system.Favorite{}).Select("video_id").Where("user_id = ? and status = ?", userId, 1).Find(&videoIds).Error
	if err != nil {
		return nil, err
	}
	var videoList *[]system.Video
	err = global.DouYinDB.Model(system.Video{}).Where("video_id IN (?)", videoIds).Find(&videoList).Error
	if err != nil {
		return nil, err
	}
	user, err := ServiceGroupApp.UserService.GetUserInfoById(userId)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	res := make([]comRes.Video, 0, len(*videoList))
	for i := 0; i < len(*videoList); i++ {
		wg.Add(1)
		go func(videoSys system.Video) {
			defer wg.Done()
			isFavorite, _ := ServiceGroupApp.FavoriteService.IsFavorite(userId, videoSys.ID)
			isFollow, _ := ServiceGroupApp.RelationService.IsFollow(userId, user.ID)
			video := &comRes.Video{
				ID:    videoSys.ID,
				Title: videoSys.Title,
				Author: comRes.User{
					ID:            user.ID,
					Name:          user.UserName,
					FollowCount:   user.FollowCount,
					FollowerCount: user.FollowerCount,
					IsFollow:      isFollow,
				},
				PlayURL:       videoSys.PlayUrl,
				CoverURL:      videoSys.CoverUrl,
				CommentCount:  videoSys.CommentCount,
				FavoriteCount: videoSys.FavoriteCount,
				IsFavorite:    isFavorite,
				CreateDate:    videoSys.CreatedAt,
			}
			res = append(res, *video)
		}((*videoList)[i])
	}
	wg.Wait()
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreateDate.After(res[j].CreateDate)
	})
	return &res, err
}

//点赞
func (f *FavoriteService) FavoriteAction(userId, videoId uint64) (*comRes.Response, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()
	//查找是否有一个被取消点赞的记录
	var favorite system.Favorite
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	err := tx.Where("user_id = ? and video_id = ?", userId, videoId).First(&favorite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { //无记录
		favorite = system.Favorite{
			UserId:  userId,
			VideoId: videoId,
			Status:  1, //1-点赞
		}
		if err := tx.Create(&favorite).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			tx.Rollback() // 遇到错误时回滚事务
			return nil, err
		}
	} else if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	} else {
		if err := tx.Model(&system.Favorite{}).Where("user_id = ? and video_id = ?", userId, videoId).Update("status", 1).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			tx.Rollback() // 遇到错误时回滚事务
			return nil, err
		}
	}
	//点赞成功后，更新视频被点赞数
	if err := tx.Model(&system.Video{}).Where("video_id = ?", videoId).Update("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()
	return &comRes.Response{
		StatusCode: 0,
		StatusMsg:  "favorite action success",
	}, nil
}

//取消点赞
func (f *FavoriteService) CancelFavoriteAction(userId, videoId uint64) (*comRes.Response, error) {
	// 开始事务
	tx := global.DouYinDB.Begin()
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	if err := tx.Model(&system.Favorite{}).Where("user_id = ? and video_id = ?", userId, videoId).Update("status", 0).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//取消点赞成功后，更新视频被点赞数
	if err := tx.Model(&system.Video{}).Where("video_id = ?", videoId).Update("favorite_count", gorm.Expr("favorite_count - 1")).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return nil, err
	}
	//提交事务
	tx.Commit()
	return &comRes.Response{
		StatusCode: 0,
		StatusMsg:  "cancel favorite action success",
	}, nil
}

//判断用户是否为该视频点赞
func (f *FavoriteService) IsFavorite(userId, videoId uint64) (bool, error) {
	if userId <= 0 {
		return false, nil
	}
	var favorite system.Favorite
	err := global.DouYinDB.Model(&system.Favorite{}).Where("user_id =? and video_id = ?", userId, videoId).First(&favorite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		global.DouYinLOG.Error("check IsFavorite failed", zap.Error(err))
		return false, err
	}
	if favorite.Status == 1 { //0-未点赞，1-点赞
		return true, nil
	}
	return false, nil
}
