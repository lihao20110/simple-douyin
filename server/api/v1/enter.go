package v1

import (
	"github.com/lihao20110/simple-douyin/server/model/system"
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

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
var usersLoginInfo = map[string]system.User{}
