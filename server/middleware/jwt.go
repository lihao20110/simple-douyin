package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/utils"
)

//JWTAuth  定义jwt中间件，进行用户身份认证
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		//从请求中获取token
		token := c.Query("token")
		if token == "" {
			token = c.PostForm("token")
		}
		if token == "" { //没有token
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 401,
				StatusMsg:  "未登录状态",
			})
			c.Abort() //阻止执行
			return
		}
		// parseToken 解析token包含的信息
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 402,
				StatusMsg:  "token 不正确",
			})
			c.Abort()
			return
		}
		if claims.ExpiresAt.Unix()-time.Now().Unix() <= 0 { //token过期，是否延期换票有待和前端约定实现
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 403,
				StatusMsg:  "token 过期",
			})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID) // 后面请求解析可以通过Get()获取UserID
		c.Next()
	}
}
