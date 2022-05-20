package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/model"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type CommentApi struct {
}

// CommentAction no practical effect, just check if token is valid
func (com *CommentApi) CommentAction(c *gin.Context) {
	token := c.Query("token")
	if _, exist := usersLoginInfo[token]; exist {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 0,
		})
	} else {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 1,
			StatusMsg:  "User doesn't exist",
		})
	}
}

// CommentList all videos have same demo comment list
func (com *CommentApi) CommentList(c *gin.Context) {
	c.JSON(http.StatusOK, sysRes.CommentListResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		CommentList: model.DemoComments,
	})
}
