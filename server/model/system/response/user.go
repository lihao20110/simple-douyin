package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type UserLoginResponse struct {
	comRes.Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	comRes.Response
	User comRes.User `json:"user"`
}
