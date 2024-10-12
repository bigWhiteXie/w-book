package repo

import (
	"context"
	"strconv"

	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-common/common/sql"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// mockgen -source=internal/repo/userrepo.go -destination mocks/repo/userrepository_mock.go
type IUserRepository interface {
	TX(fun func(txRepo *UserRepository) error) error
	Create(ctx context.Context, user *model.User) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserById(ctx context.Context, id int) (*model.User, error)
	FindOrCreate(ctx context.Context, phone string) (*model.User, error)
}

type UserRepository struct {
	userDao   *dao.UserDao
	userCache cache.UserCache
	isTx      bool
}

func NewUserRepository(userDao *dao.UserDao, cache cache.UserCache) IUserRepository {
	return &UserRepository{userDao: userDao, userCache: cache}
}

func (d *UserRepository) TX(fun func(txRepo *UserRepository) error) error {
	if d.isTx {
		return fun(d)
	}
	return d.userDao.TX(func(txDao *dao.UserDao) error {
		userRepo := NewUserRepository(txDao, d.userCache).(*UserRepository)
		userRepo.isTx = true
		return fun(userRepo)
	})
}

func (d *UserRepository) Create(ctx context.Context, user *model.User) error {
	return d.userDao.Create(ctx, user)
}

func (d *UserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := d.userDao.FindOne(ctx, &model.User{Email: sql.StringToNullString(email)})
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		return nil, codeerr.WithCode(codeerr.UserEmailNotExistCode, "can't find any user by email %s", email)
	}
	return user, nil
}

func (d *UserRepository) FindUserById(ctx context.Context, id int) (*model.User, error) {
	user, err := d.userCache.Get(ctx, strconv.Itoa(id))
	if err == nil {
		return user, nil
	}

	user, err = d.userDao.FindOne(ctx, &model.User{Id: id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, codeerr.WithCode(codeerr.UserIdNotExistCode, "can't find any user by id %d", id)
		}
		return nil, err
	}

	if err = d.userCache.Set(ctx, user); err != nil {
		logx.Errorf("fail to set user [%v] cache,cause:%s", user, err)
	}
	return user, nil
}

func (d *UserRepository) FindOrCreate(ctx context.Context, phone string) (*model.User, error) {
	//根据phone查找用户
	user, err := d.userDao.FindOne(ctx, &model.User{Email: sql.StringToNullString(phone)})
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		return user, nil
	}
	//不存在则创建用户
	err = d.userDao.Create(ctx, &model.User{Phone: sql.StringToNullString(phone)})
	if err != nil && err != dao.ErrUserEmailDuplicate {
		return nil, err
	}
	//若唯一键冲突说明用户已经被创建，查找用户并返回
	return d.userDao.FindOne(ctx, &model.User{Phone: sql.StringToNullString(phone)})
}
