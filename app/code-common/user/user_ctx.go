package user

import (
	"context"
	"strconv"
)

func GetUidByCtx(ctx context.Context) int64 {
	id, ok := ctx.Value("id").(int)
	if !ok {
		if uid, ok := ctx.Value("id").(string); ok {
			id, _ = strconv.Atoi(uid)
		}
	}
	return int64(id)
}
