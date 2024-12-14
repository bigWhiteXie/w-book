package repo

import (
	"sync"

	"gorm.io/gorm"
)

type BaseRepo struct {
	db   *gorm.DB
	lock sync.RWMutex
}

func NewBaseRepo(db *gorm.DB) *BaseRepo {
	return &BaseRepo{db: db}
}

func (repo *BaseRepo) GetDB() *gorm.DB {
	repo.lock.RLock()
	defer repo.lock.RUnlock()
	return repo.db
}

func (repo *BaseRepo) SetDB(db *gorm.DB) {
	repo.lock.Lock()
	defer repo.lock.Unlock()
	repo.db = db
}
