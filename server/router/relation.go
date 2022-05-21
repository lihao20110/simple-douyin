package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
	"github.com/lihao20110/simple-douyin/server/middleware"
)

type RelationRouter struct {
}

func (r *RelationRouter) InitRelationRouter(Router *gin.RouterGroup) {
	relationRouter := Router.Group("/relation")
	relationApi := v1.ApiGroupApp.RelationApi
	{
		relationRouter.GET("/follow/list/", relationApi.FollowList)
		relationRouter.GET("/follower/list/", relationApi.FollowerList)

		relationRouter.Use(middleware.JWTAuth()).POST("/action/", relationApi.RelationAction) //鉴权
	}
}
