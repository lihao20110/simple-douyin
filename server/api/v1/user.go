package v1

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
)

type UserApi struct{}

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
var usersLoginInfo = map[string]comRes.User{
	"zhangleidouyin": {
		Id:            1,
		Name:          "zhanglei",
		FollowCount:   10,
		FollowerCount: 5,
		IsFollow:      true,
	},
}

var userIdSequence = int64(1)

func (u *UserApi) UserInfo(c *gin.Context) {
	token := c.Query("token")
	if user, exist := usersLoginInfo[token]; exist {
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{StatusCode: 0},
			User:     user,
		})
	} else {
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 1,
				StatusMsg:  "User don't exist",
			},
		})
	}
}

func (u *UserApi) Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	token := username + password
	if _, exist := usersLoginInfo[token]; exist {
		c.JSON(http.StatusOK, sysRes.UserLoginResponse{
			Response: comRes.Response{
				StatusCode: 1,
				StatusMsg:  "User alredy exist",
			},
		})
	} else {
		atomic.AddInt64(&userIdSequence, 1)
		newUser := comRes.User{
			Id:   userIdSequence,
			Name: username,
		}
		usersLoginInfo[token] = newUser
		c.JSON(http.StatusOK, sysRes.UserLoginResponse{
			Response: comRes.Response{
				StatusCode: 0,
			},
			UserId: userIdSequence,
			Token:  username + password,
		})
	}
}

func (u *UserApi) Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	token := username + password
	if user, exist := usersLoginInfo[token]; exist {
		c.JSON(http.StatusOK, sysRes.UserLoginResponse{
			Response: comRes.Response{
				StatusCode: 0,
			},
			UserId: user.Id,
			Token:  token,
		})
	} else {
		c.JSON(http.StatusOK, sysRes.UserLoginResponse{
			Response: comRes.Response{
				StatusCode: 1,
				StatusMsg:  "User don't exist",
			},
		})
	}
}
