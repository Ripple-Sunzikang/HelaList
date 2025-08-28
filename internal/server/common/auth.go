package common

import (
	"HelaList/internal/model"
	"errors"
	"log"
	"os"
	"time"

	"HelaList/configs"

	"github.com/OpenListTeam/go-cache"
	"github.com/golang-jwt/jwt/v4"
)

var SecretKey []byte

type UserClaims struct {
	Username string `json:"username"`
	PwdTS    int64  `json:"pwd_ts"`
	jwt.RegisteredClaims
}

var validTokenCache = cache.NewMemCache[bool]()

// init 函数会在包被初次加载时自动执行
func init() {
	// 最佳实践是从环境变量或配置文件中读取密钥
	// 为了方便，这里我们先使用一个固定的密钥
	// 警告：请在生产环境中替换为一个更长、更随机的安全密钥！
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-default-long-and-secure-random-string-for-dev"
		log.Println("警告: 未设置 JWT_SECRET 环境变量, 正在使用默认的开发密钥。请勿在生产环境中使用。")
	}
	SecretKey = []byte(secret)
}

func GenerateToken(user *model.User) (tokenString string, err error) {
	claim := UserClaims{
		Username: user.Username,
		PwdTS:    user.PasswordTS,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(configs.Conf.TokenExpiresIn) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err = token.SignedString(SecretKey)
	if err != nil {
		return "", err
	}
	validTokenCache.Set(tokenString, true)
	return tokenString, err
}

func ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if IsTokenInvalidated(tokenString) {
		return nil, errors.New("token is invalidated")
	}
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("that's not even a token")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token is expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("token not active yet")
			} else {
				return nil, errors.New("couldn't handle this token")
			}
		}
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("couldn't handle this token")
}

func InvalidateToken(tokenString string) error {
	if tokenString == "" {
		return nil // don't invalidate empty guest token
	}
	validTokenCache.Del(tokenString)
	return nil
}

func IsTokenInvalidated(tokenString string) bool {
	_, ok := validTokenCache.Get(tokenString)
	return !ok
}
