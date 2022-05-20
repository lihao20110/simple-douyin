package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type VideoListResponse struct {
	comRes.Response
	VideoList []comRes.Video `json:"video_list"`
}
