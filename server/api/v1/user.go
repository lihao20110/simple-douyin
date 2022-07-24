package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type UserApi struct{}

//Register 用户注册
func (u *UserApi) Register(c *gin.Context) {
	//1.获取参数
	params := sysReq.UserRequest{}
	params.Username = c.Query("username")
	params.Password = c.Query("password")

	//2.参数合法性检验
	if err := userService.IsLegal(params.Username, params.Password); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}
	//3.user service层处理
	user, err := userService.Register(params.Username, params.Password)
	//4.返回响应
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}
	//4.jwt签发生成token
	token, err := utils.CreateToken(user.ID, user.Name)
	if err != nil {
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}
	//5.创建成功后返回用户user 和 token
	c.JSON(http.StatusOK, sysRes.UserResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "User register success!",
		},
		UserID: user.ID,
		Token:  token,
	})
}

//Login 用户登录
func (u *UserApi) Login(c *gin.Context) {
	//1.获取参数
	params := sysReq.UserRequest{}
	params.Username = c.Query("username")
	params.Password = c.Query("password")

	//2.参数合法性检验
	if err := userService.IsLegal(params.Username, params.Password); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}
	//3.user service层处理
	user, err := userService.Login(params.Username, params.Password)
	//4.返回响应
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}
	//jwt生成对应的token
	token, err := utils.CreateToken(user.ID, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, comRes.Response{
			StatusCode: 500,
			StatusMsg:  err.Error(),
		})
	}
	//登录成功，返回响应信息
	c.JSON(http.StatusOK, sysRes.UserResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "User login success!",
		},
		UserID: user.ID,
		Token:  token,
	})
}

//UserInfo 用户信息
func (u *UserApi) UserInfo(c *gin.Context) {
	//1.获取参数
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user_id is invalid",
		})
		return
	}
	//2.service层处理
	var userInfo system.User
	if err := userService.GetUserInfoByUserIDRedis(userID, &userInfo); err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  err.Error(),
		})
		return
	}
	//3.返回响应结果
	var (
		isLogined   = false //用户是否登录
		loginUserID uint64
		isFollow    = false
	)
	if token := c.Query("token"); token != "" { //判断传入的token是否存在，合法
		claims, err := utils.ParseToken(token)
		if err == nil && claims.ExpiresAt.UnixMilli()/1000-time.Now().UnixMilli()/1000 > 0 { //token合法,且在有效期内
			isLogined = true
			loginUserID = claims.UserID
		} else {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusInternalServerError, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
			return
		}
	}
	if isLogined { //用户处于登录状态
		isFollow, err = relationService.GetFollowStatus(loginUserID, userID)
		if err != nil {
			global.DouYinLOG.Error(err.Error(), zap.Error(err))
			c.JSON(http.StatusInternalServerError, comRes.Response{
				StatusCode: 500,
				StatusMsg:  err.Error(),
			})
		}
	}
	c.JSON(http.StatusOK, sysRes.UserInfoResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		User: comRes.User{
			ID:            userInfo.ID,
			Name:          userInfo.Name,
			FollowCount:   userInfo.FollowCount,
			FollowerCount: userInfo.FollowerCount,
			IsFollow:      isFollow,
		},
	})
}
