package repo

import (
	"context"

	"codexie.com/w-book-code/internal/model"
	"codexie.com/w-book-code/internal/repo/cache"
	"codexie.com/w-book-code/internal/repo/dao"
	"codexie.com/w-book-common/common/codeerr"
	"github.com/pkg/errors"
)

type SmsRepo interface {
	StoreCode(ctx context.Context, key, val, script string) error
	VerifyCode(ctx context.Context, key, val, script string) error
	SaveSmsRecord(ctx context.Context, record *model.SmsSendRecord) error
	FindById(ctx context.Context, idstr string) (*model.SmsSendRecord, error)
	UpdateById(ctx context.Context, record *model.SmsSendRecord) error
}

type SmsRepoImpl struct {
	codeCache cache.CodeCache
	codeDao   *dao.CodeDao
}

func NewSmsRepo(codeCache cache.CodeCache, codeDao *dao.CodeDao) *SmsRepoImpl {
	return &SmsRepoImpl{
		codeCache: codeCache,
		codeDao:   codeDao,
	}
}

func (repo *SmsRepoImpl) StoreCode(ctx context.Context, key, val, script string) error {
	return repo.codeCache.StoreCode(ctx, key, val, script)
}

func (repo *SmsRepoImpl) VerifyCode(ctx context.Context, key, val, script string) error {
	return repo.codeCache.VerifyCode(ctx, key, val, script)
}

func (repo *SmsRepoImpl) SaveSmsRecord(ctx context.Context, record *model.SmsSendRecord) error {
	if err := repo.codeDao.Save(ctx, record); err != nil {
		return codeerr.WithCode(codeerr.CodeNotExistErr, "fail to save %v", record)
	}
	return nil
}

func (repo *SmsRepoImpl) FindById(ctx context.Context, idstr string) (record *model.SmsSendRecord, err error) {
	if record, err = repo.codeDao.FindById(ctx, idstr); err != nil {
		return nil, errors.Wrap(codeerr.WithCode(codeerr.SmsNotFoundErr, "[SmsRepoImpl_FindById]查找短信失败,Id=%s:%s", idstr, err), "")
	}

	return record, nil
}

func (repo *SmsRepoImpl) UpdateById(ctx context.Context, record *model.SmsSendRecord) error {
	if err := repo.codeDao.Update(ctx, record); err != nil {
		return errors.Wrap(codeerr.WithCode(codeerr.SmsNotFoundErr, "[SmsRepoImpl_UpdateById]修改短信记录失败,record=%v:%s", record, err), "")
	}

	return nil
}
