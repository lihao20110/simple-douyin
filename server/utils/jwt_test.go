package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestToken(t *testing.T) {
	claims := MyCustomClaims{
		UserID:     12,
		UserName:   "hao",
		BufferTime: 86400,
		// 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                                          // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(604800))), // 过期时间 7天
			Issuer:    "hao",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "userToken",
		},
	}
	fmt.Println(claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println("-------------------------")
	fmt.Println(token)
	fmt.Println("---------------------------")
	signedString, err := token.SignedString([]byte("AllYourBase"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(signedString)
}
