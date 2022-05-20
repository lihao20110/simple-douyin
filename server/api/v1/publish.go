package v1

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/model"

	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type PublishApi struct {
}

// PublishAction check token then save upload file to public directory

func (p *PublishApi) PublishAction(c *gin.Context) {
	token := c.PostForm("token")

	if _, exist := usersLoginInfo[token]; !exist {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 1,
			StatusMsg:  "User doesn't exist",
		})
		return
	}

	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	filename := filepath.Base(data.Filename)
	user := usersLoginInfo[token]
	finalName := fmt.Sprintf("%d_%s", user.Id, filename)
	//视频上传后会保存到本地 public 目录中，访问时用 127.0.0.1:8080/static/video_name 即可

	saveFile := filepath.Join("./public/", finalName)
	if err := c.SaveUploadedFile(data, saveFile); err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, comRes.Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

// PublishList all users have same publish video list

func (p *PublishApi) PublishList(c *gin.Context) {
	c.JSON(http.StatusOK, sysRes.VideoListResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		VideoList: model.DemoVideos,
	})
}
