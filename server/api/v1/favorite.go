package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/model"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type FavoriteApi struct {
}

// FavoriteAction no practical effect, just check if token is valid

func (f *FavoriteApi) FavoriteAction(c *gin.Context) {
	token := c.Query("token")

	if _, exist := usersLoginInfo[token]; exist {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
		})
	}
	c.JSON(http.StatusOK, comRes.Response{
		StatusCode: 1,
		StatusMsg:  "User doesn't exist",
	})
}

// FavoriteList all users have same favorite video list

func (f *FavoriteApi) FavoriteList(c *gin.Context) {
	c.JSON(http.StatusOK, sysRes.VideoListResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		VideoList: model.DemoVideos,
	})
}
