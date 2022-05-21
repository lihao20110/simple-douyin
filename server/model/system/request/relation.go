package request

type RelationRequest struct {
	Token      string `json:"token"`       //用户鉴权token
	ToUserID   uint64 `json:"to_user_id"`  //对方用户id
	ActionType uint64 `json:"action_type"` //1-关注，2-取消关注
}
