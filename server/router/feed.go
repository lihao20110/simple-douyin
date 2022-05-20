package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
)

type FeedRouter struct{}

func (f *FeedRouter) InitFeedRouter(Router *gin.RouterGroup) {
	feedRouterApi := v1.ApiGroupApp.FeedApi
	Router.GET("/feed/", feedRouterApi.GetFeed)
}
