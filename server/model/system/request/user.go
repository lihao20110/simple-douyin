package request

type UserRequest struct {
	Username string `json:"username"` //登录或注册用户名
	Password string `json:"password"` //登录或注册密码
}
