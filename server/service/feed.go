package service

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type FeedService struct {
}

// 获取视频流Feed:不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func (f *FeedService) FeedVideoList(startTime string, token string) (*[]comRes.Video, string, error) {
	if startTime == "" || startTime == "0" {
		startTime = time.Now().Format("2006-01-02 15:04:05") // 不传默认为当前服务器时间
	} else {
		lastTime, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			startTime = time.Now().Format("2006-01-02 15:04:05") // 解析出错默认为当前服务器时间
		}
		startTime = time.Unix(lastTime/1000, 0).Format("2006-01-02 15:04:05")
	}
	res := make([]comRes.Video, 0, global.FeedVideoNum)
	nextTime := string(time.Now().Unix())
	var videoList *[]system.Video
	err := global.DouYinDB.Model(&system.Video{}).Where("created_at <= ?", startTime).Order("created_at DESC").Limit(global.FeedVideoNum).Find(&videoList).Error
	if err != nil {
		return nil, "", err
	}
	if token != "" { //用户处于登录状态
		id := utils.GetUserId(token)
		if id == 0 {
			return nil, "", errors.New("get userId by token failed! ")
		}
		wg := sync.WaitGroup{}
		for i := 0; i < len(*videoList); i++ {
			wg.Add(1)
			go func(videoSys system.Video) {
				defer wg.Done()
				user, _ := ServiceGroupApp.UserService.GetUserInfoById(videoSys.UserId)
				isFavorite, _ := ServiceGroupApp.FavoriteService.IsFavorite(id, videoSys.ID)
				isFollow, _ := ServiceGroupApp.RelationService.IsFollow(id, user.ID)
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
				}
				res = append(res, *video)
			}((*videoList)[i])
		}
		if len(*videoList) == global.FeedVideoNum {
			nextTime = string((*videoList)[len(*videoList)-1].CreatedAt.Unix())
		}
		wg.Wait()
		sort.Slice(res, func(i, j int) bool {
			return res[i].CreateDate.After(res[j].CreateDate)
		})
		return &res, nextTime, err
	}
	//用户处于未登录状态
	wg := sync.WaitGroup{}
	for i := 0; i < len(*videoList); i++ {
		wg.Add(1)
		go func(videoSys system.Video) {
			defer wg.Done()
			user, _ := ServiceGroupApp.UserService.GetUserInfoById(videoSys.UserId)

			video := &comRes.Video{
				ID:    videoSys.ID,
				Title: videoSys.Title,
				Author: comRes.User{
					ID:            user.ID,
					Name:          user.UserName,
					FollowCount:   user.FollowCount,
					FollowerCount: user.FollowerCount,
					IsFollow:      false, //未登录，默认未关注
				},
				PlayURL:       videoSys.PlayUrl,
				CoverURL:      videoSys.CoverUrl,
				CommentCount:  videoSys.CommentCount,
				FavoriteCount: videoSys.FavoriteCount,
				IsFavorite:    false, //未登录，默认未点赞
			}
			res = append(res, *video)
		}((*videoList)[i])
	}
	if len(*videoList) == global.FeedVideoNum {
		nextTime = string((*videoList)[len(*videoList)-1].CreatedAt.Unix())
	}
	wg.Wait()
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreateDate.After(res[j].CreateDate)
	})
	return &res, nextTime, err
}
