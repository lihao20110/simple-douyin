package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/service"
	"go.uber.org/zap"
)

type FeedApi struct{}

// Feed same demo video list for every request
//视频上传后会保存到本地 public 目录中，访问时用 127.0.0.1:8080/static/video_name 即可

func (f *FeedApi) GetFeed(c *gin.Context) {
	//1.获取请求参数
	params := sysReq.FeedRequest{}
	params.LatestTime = c.Query("latest_time")
	params.Token = c.Query("token")
	//2.service层处理

	feedVideoList, nextTime, err := service.ServiceGroupApp.FeedService.FeedVideoList(params.LatestTime, params.Token)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 402,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应
	if len(*feedVideoList) == 0 { //刷取新视频数为0
		global.DouYinLOG.Info("刷取新视频数为0", zap.String("刷取新视频数为0", "刷取新视频数为0"))
		c.JSON(http.StatusOK, sysRes.FeedResponse{
			Response: comRes.Response{
				StatusCode: 0,
				StatusMsg:  "刷取新视频数为0",
			},
			VideoList: []comRes.Video{},
			NextTime:  time.Now().Unix(),
		})
		return
	}

	next, _ := strconv.ParseInt(nextTime, 10, 64)
	c.JSON(http.StatusOK, sysRes.FeedResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "get feed video success",
		},
		VideoList: *feedVideoList,
		NextTime:  next,
	})
}
