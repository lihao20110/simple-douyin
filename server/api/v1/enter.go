package v1

import (
	"github.com/lihao20110/simple-douyin/server/service"
)

type ApiGroup struct {
	FeedApi
	UserApi
	PublishApi
	FavoriteApi
	CommentApi
	RelationApi
}

var ApiGroupApp = new(ApiGroup)

var (
	feedService     = service.ServiceGroupApp.FeedService
	userService     = service.ServiceGroupApp.UserService
	publishService  = service.ServiceGroupApp.PublishService
	favoriteService = service.ServiceGroupApp.FavoriteService
	relationService = service.ServiceGroupApp.RelationService
	commentService  = service.ServiceGroupApp.CommentService
)
