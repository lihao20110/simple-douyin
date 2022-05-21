package response

type User struct {
	ID            uint64 `json:"id,omitempty"`             // 用户id
	Name          string `json:"name,omitempty"`           // 用户名称
	FollowCount   uint64 `json:"follow_count,omitempty"`   // 关注总数
	FollowerCount uint64 `json:"follower_count,omitempty"` // 粉丝总数
	IsFollow      bool   `json:"is_follow,omitempty"`      // true-已关注，false-未关注

}
