package ijwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	TokenKey   = []byte("3zDAzPwR9s66FMlJRWKzzK5WJwvKrBPS")
	RefreshKey = []byte("3zDAzPwRfdsfasdfaf134234413")
)

var (
	TokenValidErr = errors.New("token解析错误")
	SidLogoutErr  = errors.New("token已经注销")
	UserAgentErr  = errors.New("user-agent不一致")
)
var rcmd redis.Cmdable

func InitJwtHandler(redis redis.Cmdable) {
	rcmd = redis
}

func ParseJWTToken(tokenString string, secret []byte) (*UserClaims, error) {
	claim := &UserClaims{}
	// 解析令牌
	if _, err := jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}); err != nil {
		return nil, TokenValidErr
	}
	return claim, nil
}
func SetLoginJWTToken(w http.ResponseWriter, r *http.Request, uid int) error {
	//sid标识用户token是否有效，redis中存在表示无效
	sid, err := uuid.NewUUID()
	if err != nil {
		return errors.New("生成sid失败导致token无法生成")
	}
	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
		Uid:        uid,
		Ssid:       sid.String(),
		UserAgrent: r.Header.Get("User-Agent"),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := token.SignedString(TokenKey)

	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	refreshToken, err := refresh.SignedString(RefreshKey)

	if err != nil {
		return err
	}
	w.Header().Set("x-jwt-token", accessToken)
	w.Header().Set("x-jwt-refresh-token", refreshToken)
	return nil
}

func SetStateJWTToken(w http.ResponseWriter, r *http.Request, state string, secure bool, httponly bool) error {
	claims := &WechatClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	accessToken, err := token.SignedString(TokenKey)
	if err != nil {
		return err
	}
	cookie := http.Cookie{
		Name:     "state-token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(3 * time.Minute), // 设置 cookie 过期时间为 24 小时
		HttpOnly: true,                            // 仅允许 HTTP 协议访问，防止 JavaScript 访问
	}
	http.SetCookie(w, &cookie)
	return nil
}

func CheckTokenValid(r *http.Request, tokenString string, secret []byte) (*http.Request, error) {
	var (
		claim *UserClaims
		err   error
	)
	if claim, err = ParseJWTToken(tokenString, secret); err != nil || claim.UserAgrent != r.Header.Get("User-Agent") {
		if err == nil {
			return r, UserAgentErr
		} else {
			return r, err
		}
	}
	if err != nil {
		return r, err
	}
	// sid存在表示token失效
	SidKey := fmt.Sprintf("user:ssid:%s", claim.Ssid)
	if res, err := rcmd.Exists(context.Background(), SidKey).Result(); err != nil {
		return r, err
	} else if res == 1 {
		return r, SidLogoutErr
	}
	ctx := context.WithValue(r.Context(), "id", claim.ID)
	ctx = context.WithValue(ctx, "sid", claim.Ssid)

	return r.WithContext(ctx), nil
}

func ClearToken(sid string) error {
	SidKey := fmt.Sprintf("user:ssid:%s", sid)
	if err := rcmd.Set(context.Background(), SidKey, "", 7*24*time.Hour).Err(); err != nil {
		return err
	}
	return nil
}

type UserClaims struct {
	//Claims自带预定义的字段   相当于继承了这个结构体
	jwt.RegisteredClaims
	//声明你自己要放进去的数据
	Uid        int
	Ssid       string
	UserAgrent string
}

type WechatClaims struct {
	//Claims自带预定义的字段   相当于继承了这个结构体
	jwt.RegisteredClaims
	//声明你自己要放进去的数据
	State string
}
