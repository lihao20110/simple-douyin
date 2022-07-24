package service

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FavoriteService struct {
}

//GetFavoriteVideoListByUserIDSql 查询用户的点赞列表
func (f *FavoriteService) GetFavoriteVideoListByUserIDSql(userID uint64, videoList *[]system.Video) error {
	var favoriteVideoList []system.Favorite
	var status = true
	if err := global.DouYinDB.Where("user_id = ? and is_favorite = ?", userID, &status).Find(&favoriteVideoList).Error; err != nil {
		return err
	}
	if len(favoriteVideoList) == 0 { //点赞视频数为0，直接返回
		return nil
	}
	var (
		listZ       = make([]*redis.Z, 0, len(favoriteVideoList))
		videoIDList = make([]uint64, 0, len(favoriteVideoList))
	)
	for _, favorite := range favoriteVideoList {
		videoIDList = append(videoIDList, favorite.VideoID)
		listZ = append(listZ, &redis.Z{
			Score:  float64(favorite.UpdatedAt.UnixMilli() / 1000),
			Member: favorite.VideoID,
		})
	}
	//根据视频ID列表从缓存中查找视频信息
	if err := ServiceGroupApp.FeedService.GetVideoListByVideoIDListRedis(videoList, videoIDList); err != nil {
		return err
	}
	//将用户点赞过的视频ID列表加入缓存
	if err := f.SetFavoriteVideoIDListRedis(userID, listZ...); err != nil {
		return err
	}
	return nil
}

//AddFavoriteAction 点赞操作
func (f *FavoriteService) AddFavoriteAction(userID, videoID uint64) error {
	//查缓存
	favoriteStatus, err := f.GetVideoFavoriteStatusByUserIDRedis(userID, videoID)
	if err == nil && favoriteStatus { //缓存中存在已点赞记录
		return errors.New("重复点赞")
	}
	//查找是否有一个被取消点赞的记录或者无记录
	var favorite system.Favorite
	err = global.DouYinDB.Where("user_id = ? and video_id = ?", userID, videoID).First(&favorite).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && (*favorite.IsFavorite) == true { //已点赞过，不重复点赞
		return errors.New("重复点赞")
	}
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		if errors.Is(err, gorm.ErrRecordNotFound) { //无记录
			id, err := global.DouYinIDGenerator.NextID()
			if err != nil {
				return errors.New("id generate failed")
			}
			var status = true
			favorite = system.Favorite{
				ID:         id,
				UserID:     userID,
				VideoID:    videoID,
				IsFavorite: &status, //点赞
			}
			if err := tx.Create(&favorite).Error; err != nil {
				global.DouYinLOG.Error(err.Error(), zap.Error(err))
				tx.Rollback() // 遇到错误时回滚事务
				return err
			}
		} else {
			var status = true
			if err := tx.Model(&system.Favorite{}).Select("is_favorite").Where("user_id = ? and video_id = ?", userID, videoID).Update("is_favorite", &status).Error; err != nil {
				global.DouYinLOG.Error(err.Error(), zap.Error(err))
				return err
			}
		}
		//删除点赞视频为空值的处理
		keyEmpty := fmt.Sprintf(utils.FavoriteEmptyPattern, userID)
		global.DouYinRedis.Del(global.DouYinCONTEXT, keyEmpty)
		//点赞成功后，如果缓存存在,更新视频被点赞数,用户点赞列表
		f.AddVideoFavoriteRedis(userID, videoID)
		return nil
	})
}

//CancelFavoriteAction 取消点赞
func (f *FavoriteService) CancelFavoriteAction(userID, videoID uint64) error {
	//查缓存
	favoriteStatus, err := f.GetVideoFavoriteStatusByUserIDRedis(userID, videoID)
	if err == nil && !favoriteStatus { //缓存中不存在已点赞记录
		return errors.New("重复取消点赞")
	}
	//查找是否有一个被点赞的记录或无记录
	var favorite system.Favorite
	err = global.DouYinDB.Where("user_id = ? and video_id = ?", userID, videoID).First(&favorite).Error
	if err != nil { //执行出错或者无记录
		return err
	}
	if err == nil && (*favorite.IsFavorite) == false { //已取消点赞过，不重复取消点赞
		return errors.New("重复取消点赞")
	}
	return global.DouYinDB.Transaction(func(tx *gorm.DB) error {
		var status = false
		if err := tx.Model(&system.Favorite{}).Select("is_favorite").Where("user_id = ? and video_id = ?", userID, videoID).Update("is_favorite", &status).Error; err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			return err
		}
		//取消点赞成功后，如果缓存存在,更新视频被点赞数,用户点赞列表
		f.SubVideoFavoriteRedis(userID, videoID)
		return nil
	})
}

//GetFavoriteStatusList 获取用户对一批视频的点赞状态
func (f *FavoriteService) GetFavoriteStatusList(userID uint64, videoIDList []uint64) ([]bool, error) {
	status := make([]bool, len(videoIDList))
	var favoriteVideoIDList []uint64
	if err := f.GetFavoriteVideoIDListRedis(userID, &favoriteVideoIDList); err != nil {
		return nil, err
	}
	if len(favoriteVideoIDList) == 0 { //点赞列表为空
		return status, nil
	}
	mapFavoriteID := make(map[uint64]struct{}, 0)
	for _, v := range favoriteVideoIDList {
		mapFavoriteID[v] = struct{}{}
	}
	for i, v := range videoIDList {
		if _, ok := mapFavoriteID[v]; ok {
			status[i] = true
		}
	}
	return status, nil
}
