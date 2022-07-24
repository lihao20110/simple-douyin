package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lihao20110/simple-douyin/server/global"
)

type MyCustomClaims struct {
	UserID     uint64
	UserName   string
	BufferTime int64
	jwt.RegisteredClaims
}

//CreateToken 创建一个token
func CreateToken(userID uint64, userName string) (string, error) {
	claims := MyCustomClaims{
		UserID:     userID,
		UserName:   userName,
		BufferTime: global.DouYinCONFIG.JWT.BufferTime,
		// 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                                                                       // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(global.DouYinCONFIG.JWT.ExpiresTime))), // 过期时间 7天
			Issuer:    global.DouYinCONFIG.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "userToken",
		},
	}
	SigningKey := []byte(global.DouYinCONFIG.JWT.SigningKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SigningKey)
}

//ParseToken 解析 token
func ParseToken(tokenString string) (*MyCustomClaims, error) {
	SigningKey := []byte(global.DouYinCONFIG.JWT.SigningKey)
	token, _ := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return SigningKey, nil
	})
	if claims, ok := token.Claims.(*MyCustomClaims); ok { //&& token.Valid
		return claims, nil
	}
	return nil, errors.New("fail token")
}
