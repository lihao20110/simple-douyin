package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
)

type PublishRouter struct{}

func (p *PublishRouter) InitPublishRouter(Router *gin.RouterGroup) {
	publishRouter := Router.Group("/publish")
	publishApi := v1.ApiGroupApp.PublishApi
	{
		publishRouter.POST("/action/", publishApi.PublishAction)
		publishRouter.GET("/list/", publishApi.PublishList)
	}
}
