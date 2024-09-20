package cache

import (
	"context"

	"codexie.com/w-book-user/internal/model"
)

type UserCache interface {
	Get(ctx context.Context, id string) (*model.User, error)
	Set(ctx context.Context, user *model.User) error
}
