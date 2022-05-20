package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/model"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type RelationApi struct {
}

// RelationAction no practical effect, just check if token is valid
func (r *RelationApi) RelationAction(c *gin.Context) {
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

// FollowList all users have same follow list
func (r *RelationApi) FollowList(c *gin.Context) {
	c.JSON(http.StatusOK, sysRes.UserListResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		UserList: []comRes.User{
			model.DemoUser,
		},
	})
}

// FollowerList all users have same follower list
func (r *RelationApi) FollowerList(c *gin.Context) {
	c.JSON(http.StatusOK, sysRes.UserListResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		UserList: []comRes.User{
			model.DemoUser,
		},
	})
}
