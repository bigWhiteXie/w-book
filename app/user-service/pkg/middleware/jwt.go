package middleware

import (
	"net/http"
	"strings"

	"codexie.com/w-book-user/pkg/ijwt"
	"github.com/zeromicro/go-zero/core/logx"
)

type JwtMiddleware struct {
	AllowPath []string
}

// NewCorsMiddleware 新建跨域请求处理中间件
func NewJwtMiddleware() *JwtMiddleware {
	path := []string{"/v1/user/login", "/v1/user/refresh", "/v1/user/sign", "/v1/user/login_sms", "/v1/user/login_sms/code"}
	return &JwtMiddleware{AllowPath: path}
}

func (m *JwtMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//直接放行
		if m.Allow(r) {
			next(w, r)
			return
		}

		//校验jwt
		tokenString := r.Header.Get("Authorization")
		r, err := ijwt.CheckTokenValid(r, tokenString, ijwt.TokenKey)
		if err != ijwt.TokenValidErr && err != ijwt.SidLogoutErr {
			next(w, r)
		} else {
			logx.WithContext(r.Context()).Errorf("token认证失败，cause:%s", err)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func (m *JwtMiddleware) Allow(r *http.Request) bool {
	path := r.URL.Path
	for _, allowPath := range m.AllowPath {
		if strings.EqualFold(path, allowPath) {
			return true
		}
	}
	return false
}
