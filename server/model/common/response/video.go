package response

import (
	"time"
)

type Video struct {
	ID            uint64    `json:"id,omitempty"`             // 视频唯一标识
	Title         string    `json:"title,omitempty"`          // 视频标题
	Author        User      `json:"author,omitempty"`         // 视频作者信息
	PlayURL       string    `json:"play_url,omitempty"`       // 视频播放地址
	CoverURL      string    `json:"cover_url,omitempty"`      // 视频封面地址
	FavoriteCount uint64    `json:"favorite_count,omitempty"` // 视频的点赞总数
	CommentCount  uint64    `json:"comment_count,omitempty"`  // 视频的评论总数
	IsFavorite    bool      `json:"is_favorite,omitempty"`    // true-已点赞，false-未点赞
	CreateDate    time.Time `json:"create_date"`              //视频发布时间
}
