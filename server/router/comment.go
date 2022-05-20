package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/lihao20110/simple-douyin/server/api/v1"
)

type CommentRouter struct {
}

func (c *CommentRouter) InitCommentRouter(Router *gin.RouterGroup) {
	commentRouter := Router.Group("/comment")
	commentApi := v1.ApiGroupApp.CommentApi
	{
		commentRouter.POST("/action/", commentApi.CommentAction)
		commentRouter.GET("/List", commentApi.CommentList)
	}
}
