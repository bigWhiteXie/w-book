package repo

import "context"

type CodeCache interface {
	StoreCode(ctx context.Context, key, val, script string) error
	VerifyCode(ctx context.Context, key, val, script string) error
}
