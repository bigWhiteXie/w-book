package repo

import (
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/pkg/common/code"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
)

type UserRepository struct {
	userDao   *dao.UserDao
	userCache *cache.UserCache
	isTx      bool
}

func NewUserRepository(userDao *dao.UserDao, cache *cache.UserCache) *UserRepository {
	return &UserRepository{userDao: userDao, userCache: cache}
}

func (d *UserRepository) TX(fun func(txRepo *UserRepository) error) error {
	if d.isTx {
		return fun(d)
	}
	return d.userDao.TX(func(txDao *dao.UserDao) error {
		userRepo := NewUserRepository(txDao, d.userCache)
		userRepo.isTx = true
		return fun(userRepo)
	})
}

func (d *UserRepository) Create(ctx context.Context, user *model.User) error {
	return d.userDao.Create(ctx, user)
}

func (d *UserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	users, err := d.userDao.Find(ctx, &model.User{Email: email})
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if len(users) > 0 {
		return &users[0], nil
	}
	return nil, codeerr.WithCode(code.UserEmailNotExistCode, "can't find any user by email %s", email)
}

func (d *UserRepository) FindUserById(ctx context.Context, id int) (*model.User, error) {
	user, err := d.userCache.Get(ctx, strconv.Itoa(id))
	if err == nil {
		return user, nil
	}

	users, err := d.userDao.Find(ctx, &model.User{Id: id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, codeerr.WithCode(code.UserIdNotExistCode, "can't find any user by id %d", id)
		}
		return nil, err
	}

	if err := d.userCache.Set(ctx, &users[0]); err != nil {
		logx.Errorf("fail to set user cache,%v", user)
	}
	return &users[0], nil
}
