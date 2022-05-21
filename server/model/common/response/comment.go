package response

type Comment struct {
	ID         uint64 `json:"id,omitempty" `         // 评论id
	User       User   `json:"user" gorm:"embedded"`  // 评论用户信息
	Content    string `json:"content,omitempty"`     // 评论内容
	CreateDate string `json:"create_date,omitempty"` // 评论发布日期，格式 mm-dd
}
