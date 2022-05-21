package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
	"github.com/lihao20110/simple-douyin/server/middleware"
)

type CommentRouter struct {
}

func (c *CommentRouter) InitCommentRouter(Router *gin.RouterGroup) {
	commentRouter := Router.Group("/comment")
	commentApi := v1.ApiGroupApp.CommentApi
	{
		commentRouter.GET("/list/", commentApi.CommentList)

		commentRouter.Use(middleware.JWTAuth()).POST("/action/", commentApi.CommentAction) //鉴权
	}
}
