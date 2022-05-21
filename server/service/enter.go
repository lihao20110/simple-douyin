package service

type ServiceGroup struct {
	UserService
	FeedService
	RelationService
	FavoriteService
	CommentService
	PublishService
}

var ServiceGroupApp = new(ServiceGroup)
