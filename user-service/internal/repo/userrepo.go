package repo

import (
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo/dao"
	"context"
)

type UserRepository struct {
	userDao *dao.UserDao
	isTx    bool
}

func NewUserRepository(userDao *dao.UserDao) *UserRepository {
	return &UserRepository{userDao: userDao}
}

func (d *UserRepository) TX(fun func(txRepo *UserRepository) error) error {
	if d.isTx {
		return fun(d)
	}
	return d.userDao.TX(func(txDao *dao.UserDao) error {
		userRepo := NewUserRepository(txDao)
		userRepo.isTx = true
		return fun(userRepo)
	})
}

func (d *UserRepository) Create(ctx context.Context, user *model.User) error {
	return d.userDao.Create(ctx, user)
}

func (d *UserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	users, err := d.userDao.Find(ctx, &model.User{Email: email})
	if err != nil {
		return nil, err
	}
	return &users[0], nil
}
