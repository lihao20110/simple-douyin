package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type FeedResponse struct {
	comRes.Response
	VideoList []comRes.Video `json:"video_list"` //视频列表
	NextTime  int64          `json:"next_time"`  //本次返回视频中，发布最早时间，作为下次请求时的latest_time
}
