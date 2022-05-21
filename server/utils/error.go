package utils

import (
	"errors"
)

var (
	ErrorUserNameNull  = errors.New("用户名不合法")
	ErrorUserExit      = errors.New("用户名已注册过")
	ErrorUserNotExit   = errors.New("用户名不存在")
	ErrorPasswordNull  = errors.New("密码不合法,要求以字母开头，长度在6~18之间，只能包含字符、数字和下划线")
	ErrorPasswordError = errors.New("密码错误")
)
