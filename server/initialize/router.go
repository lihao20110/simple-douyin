package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/lihao20110/simple-douyin/server/router"
)

//初始化总路由

func InitRouters() *gin.Engine {
	Router := gin.Default()
	systemRouter := router.RouterGroupApp
	PrivateGroup := Router.Group("/douyin")
	{
		// basic apis
		systemRouter.InitFeedRouter(PrivateGroup)
		systemRouter.InitUserRouter(PrivateGroup)
		systemRouter.InitPublishRouter(PrivateGroup)
		// extra apis - I
		systemRouter.InitFavoriteRouter(PrivateGroup)
		systemRouter.InitCommentRouter(PrivateGroup)
		// extra apis - II
		systemRouter.InitRelationRouter(PrivateGroup)
	}
	return Router
}
