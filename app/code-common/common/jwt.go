package common

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func GetJwtToken(secretKey string, seconds int64, payload map[string]interface{}) (string, error) {
	claims := make(jwt.MapClaims)

	iat := time.Now().Unix()
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	for k, v := range payload {
		claims[k] = v
	}
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
