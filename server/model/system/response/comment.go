package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type CommentListResponse struct {
	comRes.Response
	CommentList []comRes.Comment `json:"comment_list,omitempty"`
}
