package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"go.uber.org/zap"

	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type PublishApi struct {
}

// PublishAction check token then save upload file to public directory
func (p *PublishApi) PublishAction(c *gin.Context) {
	//1.获取请求参数
	userId := c.GetUint64("user_id")
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
	//本地保存
	playUrl, coverUrl, err = publishService.LocalSave(c, data, userId)
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
	publishVideo := &system.Video{
		Title:    title,
		UserId:   userId,
		PlayUrl:  playUrl,
		CoverUrl: coverUrl,
	}
	//5.返回响应
	if err := publishService.CreateVideo(publishVideo); err != nil {
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

// PublishList user publish video list
//用户的视频发布列表，直接列出用户所有投稿过的视频
func (p *PublishApi) PublishList(c *gin.Context) {
	//1.获取请求参数
	userIdStr := c.Query("user_id")
	if userIdStr == "" || userIdStr == "0" {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id不能为空",
		})
		return
	}
	//2.service层处理
	userId, _ := strconv.ParseUint(userIdStr, 10, 64)
	videoList, err := publishService.GetVideoList(userId)
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "get publish video list failed",
		})
		return
	}
	//3.返回响应
	if len(*videoList) == 0 { //提示作用
		global.DouYinLOG.Info("发布作品作品为0", zap.String("发布作品作品为0", "发布作品作品为0"))
	}
	c.JSON(http.StatusOK, sysRes.PublishListResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  fmt.Sprintf("共%d条发布", len(*videoList)),
		},
		VideoList: *videoList,
	})
}
