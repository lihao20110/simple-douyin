package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/router"
)

//初始化总路由

func InitRouters() *gin.Engine {
	Router := gin.Default()
	// public directory is used to serve static resources
	Router.Static("/static", "./public")
	systemRouter := router.RouterGroupApp
	PrivateGroup := Router.Group("/douyin")
	{
		// basic apis
		systemRouter.InitFeedRouter(PrivateGroup) // 视频流基础功能路由注册
		systemRouter.InitUserRouter(PrivateGroup) //用户注册、登录、用户信息功能路由注册

		systemRouter.InitPublishRouter(PrivateGroup) //发布视频，视频列表功能路由注册
		// extra apis - I
		systemRouter.InitFavoriteRouter(PrivateGroup) //点赞视频，点赞列表功能路由注册
		systemRouter.InitCommentRouter(PrivateGroup)  //发评论、评论列表功能路由注册
		// extra apis - II
		systemRouter.InitRelationRouter(PrivateGroup) //关注列表、粉丝列表功能路由注册
	}
	return Router
}
