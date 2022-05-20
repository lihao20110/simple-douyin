package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
)

type UserRouter struct{}

func (u *UserRouter) InitUserRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("/user")
	userApi := v1.ApiGroupApp.UserApi
	{
		userRouter.GET("/", userApi.UserInfo)
		userRouter.POST("/register/", userApi.Register)
		userRouter.POST("/login/", userApi.Login)
	}
}
