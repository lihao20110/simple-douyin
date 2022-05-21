package request

import (
	comReq "github.com/lihao20110/simple-douyin/server/model/common/request"
)

type CommentRequest struct {
	comReq.UserTokenRequest
	ActionType  uint   `json:"action_type"`            //1-发布评论，2-删除评论
	CommentText string `json:"comment_text,omitempty"` //用户填写的评论内容，在action_type=1的时候使用
	CommentID   uint64 `json:"comment_id,omitempty"`   //要删除的评论id，在action_type=2的时候使用
}

// CommentListRequest 评论列表的请求
type CommentListRequest struct {
	UserID  uint64 `form:"user_id" json:"user_id"`
	Token   string `form:"token" json:"token"`
	VideoID uint64 `form:"video_id" json:"video_id"`
}
