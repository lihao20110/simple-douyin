package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type FavoriteApi struct {
}

func (f *FavoriteApi) FavoriteAction(c *gin.Context) {
	//1.获取请求参数
	videoIDStr := c.Query("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "video_id is invalid",
		})
		return
	}
	actionTypeStr := c.Query("action_type")
	if actionTypeStr != "1" && actionTypeStr != "2" {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "action_type is not exist",
		})
		return
	}
	userID := c.GetUint64("user_id")
	//2.service层处理，返回响应
	if actionTypeStr == "1" { //1-点赞，
		if err := favoriteService.AddFavoriteAction(userID, videoID); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
			StatusMsg:  "add favorite action success",
		})
		return
	} else { //2-取消点赞
		if err := favoriteService.CancelFavoriteAction(userID, videoID); err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
			StatusMsg:  "cancel favorite action success",
		})
		return
	}
}

func (f *FavoriteApi) FavoriteList(c *gin.Context) {
	//1.获取请求参数
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id不合法",
		})
		return
	}
	//2.service层处理
	var videoList []system.Video
	//得到用户发布过的视频列表
	if err = favoriteService.GetFavoriteVideoListRedis(userID, &videoList); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "get favorite video list failed",
		})
		return
	}
	//3.返回响应
	if len(videoList) == 0 {
		global.DouYinLOG.Info("目前点赞视频数为0", zap.Any(userIDStr, "favorite"))
		c.JSON(http.StatusOK, sysRes.PublishListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "favorite num is 0",
			},
			VideoList: nil,
		})
	}
	var (
		videoResList   []comRes.Video
		isLogined      = false //用户是否登录
		loginUserID    uint64
		isFavoriteList []bool
		isFavorite     = false
		isFollowList   []bool
		isFollow       = false
	)
	videoIDList := make([]uint64, len(videoList))
	authorIDList := make([]uint64, len(videoList))
	for i, video := range videoList {
		videoIDList[i] = video.ID
		authorIDList[i] = video.AuthorID
	}
	//查询视频作者信息
	var authorList []system.User
	err = userService.GetUserListByIDListRedis(&authorList, authorIDList)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 500,
			StatusMsg:  err.Error(),
		})
		return
	}
	if token := c.Query("token"); token != "" { //判断传入的token是否存在，合法
		claims, err := utils.ParseToken(token)
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
		//批量获取用户对该视频的点赞状态
		isFavoriteList, err = favoriteService.GetFavoriteStatusList(loginUserID, videoIDList)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
		//批量获取用户对该视频作者的关注状态
		isFollowList, err = relationService.GetFollowStatusList(loginUserID, authorIDList)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
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
		videoResList = append(videoResList, resVideo)
	}
	c.JSON(http.StatusOK, sysRes.PublishListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get favorite list success",
		},
		VideoList: videoResList,
	})
}
