package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type UserListResponse struct {
	comRes.Response
	UserList []comRes.User `json:"user_list"`
}
