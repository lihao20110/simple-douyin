package service

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type PublishService struct {
}

//OssUpload 阿里云oss对象存储
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

//LocalSave 视频保存服务器本地
func (p *PublishService) LocalSave(c *gin.Context, data *multipart.FileHeader, userId uint64) (string, string, error) {
	//1.保存视频
	//视频上传后会保存到本地 public 目录中，访问时用 127.0.0.1:8080/static/video_name.mp4 即可
	newFileName := fmt.Sprintf("%d_%d_%s", time.Now().UnixMilli()/1000, userId, data.Filename)
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

//CheckVideo 参数合法性校验
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

//CreateVideo 将用户发布的视频写入数据库
func (p *PublishService) CreateVideo(video *system.Video) error {
	//写入数据库
	video.CreatedAt = time.Now()
	if err := global.DouYinDB.Create(video).Error; err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		return err
	}
	//发布成功后，回写缓存
	keyPublish := fmt.Sprintf(utils.PublishPattern, video.AuthorID)
	keyVideo := fmt.Sprintf(utils.VideoPattern, video.ID)
	keyEmpty := fmt.Sprintf(utils.PublishEmptyPattern, video.AuthorID)
	listZ := &redis.Z{
		Score:  float64(time.Now().UnixMilli() / 1000),
		Member: video.ID,
	}
	_, err := global.DouYinRedis.TxPipelined(global.DouYinCONTEXT, func(pipe redis.Pipeliner) error {
		//删除发布视频为空值的处理
		pipe.Del(global.DouYinCONTEXT, keyEmpty)
		//添加视频到feed缓存
		pipe.ZAdd(global.DouYinCONTEXT, utils.FeedPattern, &redis.Z{
			Score:  float64(video.CreatedAt.UnixMilli() / 1000),
			Member: video.ID,
		})
		//添加用户发布视频列表ID缓存,如果存在
		result, err := pipe.Expire(global.DouYinCONTEXT, keyPublish, utils.PublishExpire+utils.GetRandExpireTime()).Result()
		if err == nil && result {
			pipe.ZAdd(global.DouYinCONTEXT, keyPublish, listZ)
		}
		//设置视频详细信息缓存
		pipe.HMSet(global.DouYinCONTEXT, keyVideo, "id", video.ID, "created_at", video.CreatedAt.UnixMilli(), "author_id", video.AuthorID, "title", video.Title, "play_url", video.PlayUrl, "cover_url", video.CoverUrl,
			"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount)
		pipe.Expire(global.DouYinCONTEXT, keyVideo, utils.VideoExpire+utils.GetRandExpireTime())
		return nil
	})
	return err
}

//GetVideoListByUserIDSql 缓存不存在时，获取用户的发布视频列表
func (p *PublishService) GetVideoListByUserIDSql(userID uint64, videoList *[]system.Video) error {
	var getVideoList []system.Video
	if err := global.DouYinDB.Where("author_id = ?", userID).Find(&getVideoList).Error; err != nil {
		return err
	}
	*videoList = make([]system.Video, 0, len(getVideoList))
	//对查询结果建立map映射关系
	mapVideoIDToVideo := make(map[uint64]system.Video, len(getVideoList))
	for i, video := range getVideoList {
		var status = true
		//查询favorite_count
		if err := global.DouYinDB.Model(&system.Favorite{}).Where("video_id = ? and is_favorite = ?", video.ID, &status).Count(&getVideoList[i].FavoriteCount).Error; err != nil {
			return err
		}
		//查询comment_count
		if err := global.DouYinDB.Model(&system.Comment{}).Where("video_id = ? ", video.ID).Count(&getVideoList[i].CommentCount).Error; err != nil {
			return err
		}
		mapVideoIDToVideo[video.ID] = getVideoList[i]
	}
	for _, video := range getVideoList {
		if v, ok := mapVideoIDToVideo[video.ID]; ok {
			*videoList = append(*videoList, v)
		}
	}
	return nil
}
