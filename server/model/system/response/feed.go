package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type FeedResponse struct {
	comRes.Response
	VideoList []comRes.Video `json:"video_list,omitempty"`
	NextTime  int64          `json:"next_time,omitempty"`
}
