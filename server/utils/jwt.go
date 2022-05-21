package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lihao20110/simple-douyin/server/global"
)

type MyCustomClaims struct {
	UserId     uint64 `json:"user_id"`
	UserName   string `json:"username"`
	BufferTime int64
	jwt.RegisteredClaims
}

type JWT struct {
	SigningKey []byte
}

func NewJWT() *JWT {
	return &JWT{
		SigningKey: []byte(global.DouYinCONFIG.JWT.SigningKey),
	}
}

func (j *JWT) CreateClaims(userId uint64, userName string) MyCustomClaims {
	claims := MyCustomClaims{
		UserId:     userId,
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
	return claims
}

//创建一个token
func (j *JWT) CreateToken(claims MyCustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
	//return token.SigningString()
}

// CreateTokenByOldToken 旧token 换新token 使用归并回源singleflight,合并处理相同请求，避免并发问题
func (j *JWT) CreateTokenByOldToken(oldToken string, claims MyCustomClaims) (string, error) {
	v, err, _ := global.DouYinConcurrencyControl.Do("JWT:"+oldToken, func() (interface{}, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

// 解析 token
func (j *JWT) ParseToken(tokenString string) (*MyCustomClaims, error) {
	token, _ := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if claims, ok := token.Claims.(*MyCustomClaims); ok { //&& token.Valid
		return claims, nil
	}
	return nil, errors.New("fail token")
}
