package request

type FavoriteTokenRequest struct {
	Token      string `json:"token"`       //用户鉴权token
	VideoID    uint64 `json:"video_id"`    //视频id
	ActionType uint   `json:"action_type"` //1-点赞，2-取消点赞
}
