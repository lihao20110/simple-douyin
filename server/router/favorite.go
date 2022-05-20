package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
)

type FavoriteRouter struct{}

func (f *FavoriteRouter) InitFavoriteRouter(Router *gin.RouterGroup) {
	favoriteRouter := Router.Group("/favorite")
	favoriteApi := v1.ApiGroupApp.FavoriteApi
	{
		favoriteRouter.POST("/action/", favoriteApi.FavoriteAction)
		favoriteRouter.GET("/list", favoriteApi.FavoriteList)
	}
}
