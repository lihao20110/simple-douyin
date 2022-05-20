package router

type RouterGroup struct {
	FeedRouter
	UserRouter
	PublishRouter
	FavoriteRouter
	CommentRouter
	RelationRouter
}

var RouterGroupApp = new(RouterGroup)
