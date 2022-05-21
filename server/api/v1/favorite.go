package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"go.uber.org/zap"
)

type FavoriteApi struct {
}

func (f *FavoriteApi) FavoriteAction(c *gin.Context) {
	//1.获取请求参数
	videoIDStr := c.Query("video_id")
	videoId, err := strconv.ParseUint(videoIDStr, 10, 64)
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
	userId := c.GetUint64("user_id")
	//2.service层处理，返回响应
	if actionTypeStr == "1" { //1-点赞，
		res, err := favoriteService.FavoriteAction(userId, videoId)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "favorite action failed",
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	} else { //2-取消点赞
		res, err := favoriteService.CancelFavoriteAction(userId, videoId)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 500,
				StatusMsg:  "cancel favorite action failed",
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	}
}

func (f *FavoriteApi) FavoriteList(c *gin.Context) {
	//1.获取请求参数
	userIdStr := c.Query("user_id")
	if userIdStr == "" || userIdStr == "0" {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id不能为空",
		})
		return
	}
	userId, _ := strconv.ParseUint(userIdStr, 10, 64)
	//2.service层处理
	favoriteVideoList, err := favoriteService.FavoriteList(userId)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(*favoriteVideoList) == 0 { //点赞视频数为0
		global.DouYinLOG.Info("点赞视频数为0", zap.String("点赞视频数为0", "点赞视频数为0"))
		c.JSON(http.StatusOK, sysRes.FavoriteListResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "点赞视频数为0",
			},
			VideoList: []comRes.Video{},
		})
		return
	}
	c.JSON(http.StatusOK, sysRes.FavoriteListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get favorite list success",
		},
		VideoList: *favoriteVideoList,
	})
}
