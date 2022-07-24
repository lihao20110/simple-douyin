package utils

import (
	"math/rand"
	"time"
)

//redis中key模板

var (
	FeedPattern               = "feed"                     //所有视频的ID列表缓存(zset)的key
	VideoPattern              = "video:%d"                 //单条视频的信息缓存(hset)的key
	PublishPattern            = "publish:%d"               //单个用户的发布视频ID列表缓存(zset)的key
	UserPattern               = "user:%d"                  //单个用户的信息缓存(hset)的key
	CommentsOfVideoPattern    = "commentsOfVideo:%d"       //单条视频的所有发布评论ID列表缓存(zset)的key
	CommentPattern            = "comment:%d"               //单条评论的信息缓存(hset)的key
	FavoritePattern           = "favorite:%d"              //单个用户的所有点赞视频ID列表缓存(zset)的key
	FollowPattern             = "follow:%d"                //单个用户的所有关注用户ID列表缓存(zset)的key
	FollowerPattern           = "follower:%d"              //单个用户的所有粉丝用户ID列表缓存(zset)的key
	PublishEmptyPattern       = "publishVideoEmpty:%d"     //用户发布视频列表为空的空值处理：(set) key
	FavoriteEmptyPattern      = "favoriteVideoEmpty:%d"    //用户点赞视频列表为空的空值处理：(set) key
	FollowEmptyPattern        = "followUserEmpty:%d"       //用户关注列表为空的空值处理：(set) key
	FollowerEmptyPatten       = "userFollowerEmpty:%d"     //用户粉丝列表为空的空值处理：(set) key
	CommentEmptyPattern       = "videoCommentEmpty:%d"     //视频评论列表为空的空值处理：(set) key
	UsernameRegisteredPattern = "usernameRegistered:%s"    //用户名已被注册的缓存处理：(set) key
	UserInfoEmptyPattern      = "userInfoEmptyPattern:%d"  //用户信息查看缓存穿透的空值处理(set)key
	UserLoginEmptyPattern     = "userLoginEmptyPattern:%s" //登录用户名不存在的空值处理(set)key
)

//redis中key对应的过期时间

var (
	randExpireTime           = 3 * time.Minute
	VideoExpire              = 10 * time.Minute
	PublishExpire            = 10 * time.Minute
	UserExpire               = 10 * time.Minute
	CommentsOfVideoExpire    = 10 * time.Minute
	CommentExpire            = 10 * time.Minute
	FavoriteExpire           = 10 * time.Minute
	FollowExpire             = 10 * time.Minute
	FollowerExpire           = 10 * time.Minute
	PublishEmptyExpire       = 1 * time.Minute
	FavoriteEmptyExpire      = 1 * time.Minute
	FollowEmptyExpire        = 1 * time.Minute
	FollowerEmptyExpire      = 1 * time.Minute
	CommentEmptyExpire       = 1 * time.Minute
	UsernameRegisteredExpire = 10 * time.Second
	UserInfoEmptyExpire      = 10 * time.Second
	UserLoginEmptyExpire     = 10 * time.Second
)

//GetRandExpireTime 针对大量数据同时失效带来的缓存雪崩问题，一般的解决方案是要避免大量的数据设置相同的过期时间，如果业务上的确有要求数据要同时失效，那么可以在过期时间上加一个较小的随机数，这样不同的数据过期时间不同，但差别也不大，避免大量数据同时过期，也基本能满足业务的需求。
func GetRandExpireTime() time.Duration {
	return time.Duration(rand.Float64()*randExpireTime.Seconds()) * time.Second
}
