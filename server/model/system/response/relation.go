package response

import (
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
)

type RelationListResponse struct { //关注者、粉丝列表响应
	comRes.Response
	UserList []comRes.User `json:"user_list"` // 用户信息列表
}
