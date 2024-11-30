package db

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Article struct {
	Id       int64  `json:"",gorm:"primaryKey,autoIncrement"`
	Title    string `json:""`
	Content  string `json:"content",gorm:"type:blob"`
	AuthorId int64  `json:"",gorm:"index:idx_uid_uptime"`
	Status   uint8  `json:""`
	Ctime    int64  `json:""`
	Utime    int64  `json:"",gorm:"index:idx_uid_uptime"`
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
		return errors.Wrapf(result.Error, "[UpdateById_dsFdFEDS] 更新文章失败,authorId=%d,id=%d", artEntity.AuthorId, artEntity.Id)
	}
	if result.RowsAffected == 0 {
		return errors.Wrap(fmt.Errorf("[UpdateById_dsFdFEDS] 用户[%d]修改他人文章[%s],", artEntity.AuthorId, artEntity.Id), "")
	}
	return nil
}

func (d *AuthorDao) SelectPage(ctx context.Context, authorId int64, page, size int) ([]*Article, error) {
	var articles []*Article
	offset := (page - 1) * size
	// 执行查询：按 authorId 过滤并分页
	result := d.db.WithContext(ctx).
		Select("id", "title", "author_id", "status", "utime", "ctime").
		Where("author_id = ?", authorId).
		Order("utime DESC"). // 按创建时间倒序排列
		Limit(size).
		Offset(offset).
		Find(&articles)
	if err := result.Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrapf(err, "[AuthorDao_SelectPage] 查询文章列表失败,authorId=%d, page=%d, size=%d", authorId, page, size)
	}

	return articles, nil
}

func (d *AuthorDao) SelectById(ctx context.Context, id int64) (*Article, error) {
	article := &Article{}
	if err := d.db.Find(article, id).Error; err != nil {
		return nil, errors.Wrapf(err, "[AuthorDao_SelectById] 查询文章失败,id=%d", id)
	}
	return article, nil
}

func (d *AuthorDao) Create(ctx context.Context, artEntity *Article) error {
	now := time.Now().UnixMilli()
	artEntity.Utime = now
	artEntity.Ctime = now
	if err := d.db.Create(artEntity).Error; err != nil {
		return errors.Wrapf(err, "[AuthorDao_Create] 创建文章失败,article:%v", artEntity)
	}

	return nil
}
