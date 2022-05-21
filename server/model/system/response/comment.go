package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type CommentActionResponse struct {
	comRes.Response
	Comment comRes.Comment `json:"comment,omitempty"` // 评论成功返回评论内容，不需要重新拉取整个列表
}

type CommentListResponse struct {
	comRes.Response
	CommentList []comRes.Comment `json:"comment_list,omitempty"` // 评论列表
}
