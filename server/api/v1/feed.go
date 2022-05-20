package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/model"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type FeedApi struct{}

func (f *FeedApi) GetFeed(c *gin.Context) {
	// Feed same demo video list for every request
	c.JSON(http.StatusOK, sysRes.FeedResponse{
		Response:  comRes.Response{StatusCode: 0},
		VideoList: model.DemoVideos,
		NextTime:  time.Now().Unix(),
	})
}
