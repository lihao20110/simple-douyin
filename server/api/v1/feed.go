package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type FeedApi struct{}

//GetFeed 视频流接口,为客户端拉取视频
func (f *FeedApi) GetFeed(c *gin.Context) {
	//1.获取请求参数
	params := sysReq.FeedRequest{}
	params.LatestTime = c.Query("latest_time")
	params.Token = c.Query("token")
	//2.service层处理
	var (
		videoList  []system.Video
		authorList []system.User
	)
	if err := feedService.GetFeedVideoListRedis(&videoList, &authorList, params.LatestTime); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, sysRes.FeedResponse{
			Response: comRes.Response{
				StatusCode: 500,
				StatusMsg:  "GetFeedVideoListRedis failed",
			},
			VideoList: []comRes.Video{},
			NextTime:  time.Now().UnixMilli() / 1000,
		})
		return
	}
	if len(videoList) == 0 { //后端没有视频
		//没有的话，再以当前时间获取一遍
		if err := feedService.GetFeedVideoListRedis(&videoList, &authorList, params.LatestTime); err != nil {
			c.JSON(http.StatusOK, sysRes.FeedResponse{
				Response: comRes.Response{
					StatusCode: 500,
					StatusMsg:  "GetFeedVideoListRedis failed",
				},
				VideoList: []comRes.Video{},
				NextTime:  time.Now().UnixMilli() / 1000,
			})
			return
		}
		if len(videoList) == 0 {
			global.DouYinLOG.Info("刷取新视频数为0", zap.String("刷取新视频数为0", "刷取新视频数为0"))
			c.JSON(http.StatusOK, sysRes.FeedResponse{
				Response: comRes.Response{
					StatusCode: 0,
					StatusMsg:  "get feed video success",
				},
				VideoList: []comRes.Video{},
				NextTime:  time.Now().UnixMilli() / 1000,
			})
			return
		}
	}
	var (
		feedVideoRes   = make([]comRes.Video, 0, len(videoList))
		isLogined      = false //用户是否登录
		loginUserID    uint64
		nextTime       = time.Now().UnixMilli() / 1000
		isFavoriteList []bool
		isFavorite     = false
		isFollowList   []bool
		isFollow       = false
	)
	if params.Token != "" { //判断传入的token是否存在，合法
		claims, err := utils.ParseToken(params.Token)
		if err == nil && claims.ExpiresAt.UnixMilli()/1000-time.Now().UnixMilli()/1000 > 0 { //token合法,且在有效期内
			isLogined = true
			loginUserID = claims.UserID
		} else {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusInternalServerError, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	if isLogined { //用户处于登录状态
		//批量获取用户是否点赞了列表中的视频以及是否关注了视频的作者
		videoIDList := make([]uint64, len(videoList))
		authorIDList := make([]uint64, len(videoList))
		for i, video := range videoList {
			videoIDList[i] = video.ID
			authorIDList[i] = video.AuthorID
		}
		//批量获取用户对该视频的点赞状态
		var err error
		isFavoriteList, err = favoriteService.GetFavoriteStatusList(loginUserID, videoIDList)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, sysRes.FeedResponse{
				Response: comRes.Response{
					StatusCode: 500,
					StatusMsg:  "GetFavoriteStatusList failed",
				},
				VideoList: []comRes.Video{},
				NextTime:  time.Now().UnixMilli() / 1000,
			})
			return
		}
		//批量获取用户对该视频作者的关注状态
		isFollowList, err = relationService.GetFollowStatusList(loginUserID, authorIDList)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, sysRes.FeedResponse{
				Response: comRes.Response{
					StatusCode: 500,
					StatusMsg:  "GetFollowStatusList failed",
				},
				VideoList: []comRes.Video{},
				NextTime:  time.Now().UnixMilli() / 1000,
			})
			return
		}
	}
	//未登录状态，默认未点赞，未关注
	for i, video := range videoList {
		if isLogined {
			isFavorite = isFavoriteList[i]
			isFollow = isFollowList[i]
		}
		resVideo := comRes.Video{
			ID:    video.ID,
			Title: video.Title,
			Author: comRes.User{
				ID:            authorList[i].ID,
				Name:          authorList[i].Name,
				FollowCount:   authorList[i].FollowCount,
				FollowerCount: authorList[i].FollowerCount,
				IsFollow:      isFollow,
			},
			PlayURL:       video.PlayUrl,
			CoverURL:      video.CoverUrl,
			CommentCount:  video.CommentCount,
			FavoriteCount: video.FavoriteCount,
			IsFavorite:    isFavorite,
			CreateDate:    video.CreatedAt,
		}
		feedVideoRes = append(feedVideoRes, resVideo)
	}
	//本次返回视频中发布最早的时间,作为下一次拉取视频的分页依据
	nextTime = videoList[len(videoList)-1].CreatedAt.UnixMilli()/1000 - 1
	c.JSON(http.StatusOK, sysRes.FeedResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get feed video success",
		},
		VideoList: feedVideoRes,
		NextTime:  nextTime,
	})
}
