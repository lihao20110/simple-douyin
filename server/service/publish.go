package service

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type PublishService struct {
}

func (p *PublishService) OssUpload(data *multipart.FileHeader, video, image string) (string, string, error) {
	//1.上传视频
	objectVideo := global.DouYinCONFIG.AliyunOSS.BasePath + video
	localVideo := "./public/" + video
	var playUrl string
	var err error
	if data.Size <= global.DouYinCONFIG.AliyunOSS.PartSize*global.MB {
		playUrl, err = utils.UploadFromFile(objectVideo, localVideo)
	} else { //分片上传
		playUrl, err = utils.MultipartUpload(objectVideo, localVideo)
	}
	if err != nil {
		return "", "", err
	}
	//2.上传视频封面图片
	objectImage := global.DouYinCONFIG.AliyunOSS.BasePath + image
	localImage := "./public/" + image
	coverUrl, err := utils.UploadFromFile(objectImage, localImage)
	if err != nil {
		return "", "", err
	}
	//3.上传成功后，删除本地保存文件
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(localVideo string) {
		defer wg.Done()
		err := os.Remove(localVideo) //  "./public/1655381554_1_1001.mp4"
		if err != nil {
			global.DouYinLOG.Error(errors.New("delete local video fail").Error(), zap.Error(err))
		}
	}(localVideo)
	go func(localImage string) {
		err := os.Remove(localImage) //  "./public/1655381554_1_1001.jpeg"
		if err != nil {
			global.DouYinLOG.Error(errors.New("delete local image fail").Error(), zap.Error(err))
		}
	}(localVideo)
	wg.Done()

	return playUrl, coverUrl, nil
}
func (p *PublishService) LocalSave(c *gin.Context, data *multipart.FileHeader, userId uint64) (string, string, error) {
	//1.保存视频
	//视频上传后会保存到本地 public 目录中，访问时用 127.0.0.1:8080/static/video_name.mp4 即可
	newFileName := fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userId, data.Filename)
	videoLocalSavePath := filepath.Join("./public/", newFileName)        //工程目录server/public下
	if err := c.SaveUploadedFile(data, videoLocalSavePath); err != nil { //保存在本地
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		return "", "", err
	}
	staticIp := global.DouYinCONFIG.System.StaticIp //视频存放服务器地址:http://192.168.1.183:8080/static/
	playLocalUrl := staticIp + newFileName          //视频url

	//2.使用ffmpeg获取视频封面图片
	fileSuffix := path.Ext(newFileName)                                              //获取视频文件后缀
	imageNameOnly := strings.TrimSuffix(newFileName, fileSuffix)                     //去掉后缀
	imageLocalSavePathOnly := strings.TrimSuffix(videoLocalSavePath, fileSuffix)     //去掉后缀
	image, err := utils.GetCoverImage(videoLocalSavePath, imageLocalSavePathOnly, 1) //ffmpeg获取视频封面，并与视频同路径下保存
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		return "", "", err
	}
	global.DouYinLOG.Info(image)
	coverLocalUrl := staticIp + imageNameOnly + ".jpeg" //封面url
	return playLocalUrl, coverLocalUrl, nil
}

func (p *PublishService) CheckVideo(data *multipart.FileHeader, title string) (*comRes.Response, error) {
	//检查title
	if len(title) > global.MaxTitleLength || len(title) < global.MinTitleLength {
		return &comRes.Response{
			StatusCode: 401,
			StatusMsg:  "title length is out of limit",
		}, errors.New("title length is out of limit")
	}
	//检查文件的扩展名是否为.MP4或.mp4
	fileSuffix := path.Ext(data.Filename) //获取视频文件后缀
	if string(bytes.ToLower([]byte(fileSuffix))) != global.DouYinCONFIG.AliyunOSS.AllowExt {
		return &comRes.Response{
			StatusCode: 401,
			StatusMsg:  "视频格式不符合要求",
		}, errors.New("视频格式不符合要求")
	}
	//检查文件大小
	if data.Size > global.DouYinCONFIG.AliyunOSS.MaxSize*global.MB {
		return &comRes.Response{
			StatusCode: 401,
			StatusMsg:  "视频大小超出上传要求",
		}, errors.New("视频大小超出上传要求")
	}
	return nil, nil
}
func (p *PublishService) CreateVideo(video *system.Video) error {
	// 开始事务
	tx := global.DouYinDB.Begin()
	// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
	if err := tx.Create(video).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		tx.Rollback() // 遇到错误时回滚事务
		return err
	}
	//发布，更新发布视频数 +1   接口文档未要求返回视频发布数数据，暂待
	//提交事务
	tx.Commit()
	return nil
}

func (p *PublishService) GetVideoList(userId uint64) (*[]comRes.Video, error) {
	var videoList *[]system.Video
	err := global.DouYinDB.Where("author_id = ?", userId).Order("created_at DESC").Find(&videoList).Error
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
