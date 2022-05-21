package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
	"github.com/lihao20110/simple-douyin/server/middleware"
)

type PublishRouter struct{}

func (p *PublishRouter) InitPublishRouter(Router *gin.RouterGroup) {
	publishRouter := Router.Group("/publish")
	publishApi := v1.ApiGroupApp.PublishApi
	{
		publishRouter.GET("/list/", publishApi.PublishList)
		publishRouter.Use(middleware.JWTAuth()).POST("/action/", publishApi.PublishAction) //鉴权
	}
}
