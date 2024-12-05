package middleware

import (
	"context"
	"net/http"
)

func TestUidHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r = r.WithContext(context.WithValue(ctx, "id", 123))
		next(w, r)
	}
}
