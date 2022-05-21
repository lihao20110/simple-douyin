package request

type UserTokenRequest struct {
	UserID uint64 `json:"user_id"` //用户id
	Token  string `json:"token"`   //用户鉴权token
}
