package repo

import (
	"context"

	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-user/pkg/common/codeerr"
)

type CodeRepo interface {
	StoreCode(ctx context.Context, key, val, script string) error
	VerifyCode(ctx context.Context, key, val, script string) error
	SaveSmsRecord(ctx context.Context, record *model.SmsSendRecord) error
	FindById(ctx context.Context, idstr string) (*model.SmsSendRecord, error)
	UpdateById(ctx context.Context, record *model.SmsSendRecord) error
}

type CodeRepoImpl struct {
	codeCache cache.CodeCache
	codeDao   *dao.CodeDao
}

func NewCodeRepo(codeCache cache.CodeCache, codeDao *dao.CodeDao) *CodeRepoImpl {
	return &CodeRepoImpl{
		codeCache: codeCache,
		codeDao:   codeDao,
	}
}

func (repo *CodeRepoImpl) StoreCode(ctx context.Context, key, val, script string) error {
	return repo.codeCache.StoreCode(ctx, key, val, script)
}

func (repo *CodeRepoImpl) VerifyCode(ctx context.Context, key, val, script string) error {
	return repo.codeCache.VerifyCode(ctx, key, val, script)
}

func (repo *CodeRepoImpl) SaveSmsRecord(ctx context.Context, record *model.SmsSendRecord) error {
	if err := repo.codeDao.Save(ctx, record); err != nil {
		return codeerr.WithCode(codeerr.CodeNotExistErr, "fail to save %v", record)
	}
	return nil
}

func (repo *CodeRepoImpl) FindById(ctx context.Context, idstr string) (*model.SmsSendRecord, error) {
	return repo.codeDao.FindById(ctx, idstr)
}

func (repo *CodeRepoImpl) UpdateById(ctx context.Context, record *model.SmsSendRecord) error {
	return repo.codeDao.Update(ctx, record)
}
