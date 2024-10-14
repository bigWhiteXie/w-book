package dao

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type Article struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Title    string
	Content  string `gorm:"type:blob"`
	AuthorId int64  `gorm:"index:idx_uid_uptime"`
	Ctime    int64
	Utime    int64 `gorm:"index:idx_uid_uptime"`
}

type AuthorDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewAuthorDao(db *gorm.DB) *AuthorDao {
	return &AuthorDao{db: db}
}

func (d *AuthorDao) UpdateById(ctx context.Context, artEntity *Article) error {
	now := time.Now().UnixMilli()
	result := d.db.Model(&Article{}).Where(
		"id=? and author_id=?", artEntity.Id, artEntity.AuthorId,
	).Updates(map[string]interface{}{
		"Title":   artEntity.Title,
		"Content": artEntity.Content,
		"Utime":   now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		logx.WithContext(ctx).Errorf("[UpdateById_dsFdFEDS] 用户[%d]修改他人文章[%s],", artEntity.AuthorId, artEntity.Id)
		return errors.New("用户非法修改文章")
	}
	return nil
}

func (d *AuthorDao) Create(ctx context.Context, artEntity *Article) error {
	now := time.Now().UnixMilli()
	artEntity.Utime = now
	artEntity.Ctime = now
	return d.db.Create(artEntity).Error
}
