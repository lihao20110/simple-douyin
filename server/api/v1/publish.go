package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"

	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type PublishApi struct {
}

//PublishAction 视频投稿接口
func (p *PublishApi) PublishAction(c *gin.Context) {
	//1.获取请求参数
	userID := c.GetUint64("user_id")
	title := c.PostForm("title")
	data, err := c.FormFile("data")
	if err != nil || data == nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "get file failed from key data",
		})
		return
	}
	//2.检查请求参数
	if res, err := publishService.CheckVideo(data, title); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, *res)
		return
	}
	//3.视频保存
	var playUrl, coverUrl string
	//本地保存，视频上传后会保存到本地 public 目录中，访问时用 127.0.0.1:8080/static/video_name 即可
	playUrl, coverUrl, err = publishService.LocalSave(c, data, userID)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "视频图片保存本地失败",
		})
		return
	}
	if global.DouYinCONFIG.System.OpenOss == true { //开启对象存储
		localVideo := strings.TrimPrefix(playUrl, global.DouYinCONFIG.System.StaticIp)
		localImage := strings.TrimPrefix(coverUrl, global.DouYinCONFIG.System.StaticIp)

		playUrl, coverUrl, err = publishService.OssUpload(data, localVideo, localImage)
		if err != nil {
			global.DouYinLOG.Error("oss err", zap.Error(err))
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 401,
				StatusMsg:  "对象存储失败",
			})
			return
		}
	}
	//4.将记录保存到数据库
	videoID, err := global.DouYinIDGenerator.NextID()
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "id 生成失败",
		})
		return
	}
	publishVideo := system.Video{
		ID:       videoID,
		Title:    title,
		AuthorID: userID,
		PlayUrl:  playUrl,
		CoverUrl: coverUrl,
	}
	//5.返回响应
	if err := publishService.CreateVideo(&publishVideo); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  "publish video failed",
		})
		return
	}
	c.JSON(http.StatusOK, comRes.Response{
		StatusCode: 0,
		StatusMsg:  " uploaded successfully",
	})
}

//PublishList 用户的视频发布列表
func (p *PublishApi) PublishList(c *gin.Context) {
	//1.获取请求参数
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is error",
		})
		return
	}
	//2.service层处理
	var videoList []system.Video
	err = publishService.GetPublishedVideoListRedis(userID, &videoList) //得到用户发布过的视频
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "get publish video list failed",
		})
		return
	}
	//3.返回响应
	if len(videoList) == 0 {
		global.DouYinLOG.Info("目前发布过作品数为0", zap.Any(userIDStr, "published"))
		c.JSON(http.StatusOK, sysRes.PublishListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "published num is 0",
			},
			VideoList: nil,
		})
		return
	}
	//同一作者的视频，只需查询一次作者信息
	var author system.User
	if err := userService.GetUserInfoByUserIDRedis(userID, &author); err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 500,
			StatusMsg:  err.Error(),
		})
		return
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
		//批量获取用户是否点赞了列表中的视频以及是否关注了视频的作者
		videoIDList := make([]uint64, len(videoList))
		authorIDList := make([]uint64, len(videoList))
		for i, video := range videoList {
			videoIDList[i] = video.ID
			authorIDList[i] = video.AuthorID
		}
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
				ID:            author.ID,
				Name:          author.Name,
				FollowCount:   author.FollowCount,
				FollowerCount: author.FollowerCount,
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
			StatusMsg:  fmt.Sprintf("共%d条发布", len(videoResList)),
		},
		VideoList: videoResList,
	})
}
