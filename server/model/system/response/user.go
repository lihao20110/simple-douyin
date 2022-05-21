package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type UserResponse struct { //用户登录、注册响应
	comRes.Response
	UserID int64  `json:"user_id,omitempty"` // 用户id
	Token  string `json:"token"`             // 用户鉴权token
}

type UserInfoResponse struct { //用户信息响应
	comRes.Response
	User comRes.User `json:"user"`
}
