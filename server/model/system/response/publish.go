package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type PublishListResponse struct {
	comRes.Response
	VideoList []comRes.Video `json:"video_list"` // 用户发布的视频列表
}
