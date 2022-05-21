package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	comRes "github.com/lihao20110/simple-douyin/server/model/common/response"
	"github.com/lihao20110/simple-douyin/server/utils"
)

// JwtAuth jwt中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		//从请求中获取token
		token := c.Query("token")
		if token == "" {
			token = c.Request.PostFormValue("token")
		}
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
		userId := utils.GetUserId(token)
		if userId == 0 { //从redis中查询，token是否失效
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 403,
				StatusMsg:  "token失效",
			})
			c.Abort() //阻止执行
			return
		}
		j := utils.NewJWT()
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if err != nil || claims.UserId != userId {
			c.JSON(http.StatusOK, comRes.Response{
				StatusCode: 403,
				StatusMsg:  "token 不正确",
			})
			c.Abort()
			return
		}
		//token过期
		//换票有待和前端约定使用
		//if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime { //token换票区
		//	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(global.DouYinCONFIG.JWT.ExpiresTime))) //token延期
		//	newToken, _ := j.CreateTokenByOldToken(token, *claims)
		//	newClaims, _ := j.ParseToken(newToken)
		//	c.Header("new-token", newToken)
		//	c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
		//}

		c.Set("user_id", claims.UserId)

		c.Next()
	}
}
