package request

import (
	"mime/multipart"
)

type PublishActionRequest struct {
	Data  *multipart.FileHeader `json:"data"`  //视频数据
	Token string                `json:"token"` //用户鉴权token
	Title string                `json:"title"` //视频标题
}
