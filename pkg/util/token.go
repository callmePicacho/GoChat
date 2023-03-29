package util

import (
	"GoChat/config"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type UserClaims struct {
	UserId uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken
// 生成 token
func GenerateToken(userId uint64) (string, error) {
	UserClaim := &UserClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(config.GlobalConfig.JWT.ExpireTime))),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim)
	tokenString, err := token.SignedString([]byte(config.GlobalConfig.JWT.SignKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AnalyseToken
// 解析 token
func AnalyseToken(tokenString string) (*UserClaims, error) {
	userClaim := new(UserClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.SignKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("analyse Token Error:%v", err)
	}
	return userClaim, nil
}
