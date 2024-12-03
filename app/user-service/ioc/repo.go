package ioc

import (
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
)

func InitUserRepo(userDao *dao.UserDao, cache cache.UserCache) repo.IUserRepository {
	return repo.NewUserRepository(userDao, cache)
}
