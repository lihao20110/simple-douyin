package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
	"github.com/lihao20110/simple-douyin/server/middleware"
)

type FavoriteRouter struct{}

func (f *FavoriteRouter) InitFavoriteRouter(Router *gin.RouterGroup) {
	favoriteRouter := Router.Group("/favorite")
	favoriteApi := v1.ApiGroupApp.FavoriteApi
	{
		favoriteRouter.GET("/list/", favoriteApi.FavoriteList)

		favoriteRouter.Use(middleware.JWTAuth()).POST("/action/", favoriteApi.FavoriteAction) //鉴权
	}
}
