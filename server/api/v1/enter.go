package v1

type ApiGroup struct {
	FeedApi
	UserApi
	PublishApi
	FavoriteApi
	CommentApi
	RelationApi
}

var ApiGroupApp = new(ApiGroup)
