package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type FavoriteListResponse struct {
	comRes.Response
	VideoList []comRes.Video `json:"video_list"` // 用户点赞视频列表
}
