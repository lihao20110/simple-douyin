package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/global"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/model/system"
	sysReq "github.com/lihao20110/simple-douyin/server/model/system/request"
	sysRes "github.com/lihao20110/simple-douyin/server/model/system/response"
	"github.com/lihao20110/simple-douyin/server/service"
	"github.com/lihao20110/simple-douyin/server/utils"
	"go.uber.org/zap"
)

type UserApi struct{}

//用户注册
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
	//4.签发权限token
	token, err := u.tokenNext(c, user)
	//存储token到redis中
	err = utils.SetToken(token, user.ID)
	if err != nil {
		c.JSON(http.StatusOK, sysRes.UserResponse{
			Response: comRes.Response{
				StatusCode: 401,
				StatusMsg:  err.Error(),
			},
		})
		return
	}

	//5.创建成功后返回用户user 和 权限token
	c.JSON(http.StatusOK, sysRes.UserResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "User register success!",
		},
		UserID: int64(user.ID),
		Token:  token,
	})
}

//签发jwt
func (u *UserApi) tokenNext(c *gin.Context, user *system.User) (string, error) {
	jwt := utils.NewJWT()
	claims := jwt.CreateClaims(user.ID, user.UserName)
	token, err := jwt.CreateToken(claims)
	return token, err
}

//用户登录
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
	token, _ := u.tokenNext(c, user)
	global.DouYinLOG.Info(token, zap.String(token, token))
	//存储token到redis中
	err = utils.SetToken(token, user.ID)
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

	//登录成功，返回响应信息
	c.JSON(http.StatusOK, sysRes.UserResponse{
		Response: comRes.Response{
			StatusCode: 0,
			StatusMsg:  "User login success!",
		},
		UserID: int64(user.ID),
		Token:  token,
	})
}

//用户信息
func (u *UserApi) UserInfo(c *gin.Context) {
	//1.获取参数
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  "user id is invalid",
		})
		return
	}
	tokenUserId := c.GetUint64("user_id")
	//2.service层处理
	userInfo, err := userService.GetUserInfoById(userId)
	//3.返回响应结果
	if err != nil {
		global.DouYinLOG.Error(err.Error(), zap.Error(err))
		c.JSON(http.StatusOK, comRes.Response{
			StatusCode: 401,
			StatusMsg:  err.Error(),
		})
		return
	}
	isFollow, _ := service.ServiceGroupApp.RelationService.IsFollow(tokenUserId, userId)
	c.JSON(http.StatusOK, sysRes.UserInfoResponse{
		Response: comRes.Response{
			StatusCode: 0,
		},
		User: comRes.User{
			ID:            userInfo.ID,
			Name:          userInfo.UserName,
			FollowCount:   userInfo.FollowCount,
			FollowerCount: userInfo.FollowerCount,
			IsFollow:      isFollow,
		},
	})
}
