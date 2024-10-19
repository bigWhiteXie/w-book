package db

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
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
		return result.Error
	}
	if result.RowsAffected == 0 {
		logx.WithContext(ctx).Errorf("[UpdateById_dsFdFEDS] 用户[%d]修改他人文章[%s],", artEntity.AuthorId, artEntity.Id)
		return errors.New("用户非法修改文章")
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
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return articles, result.Error
}

func (d *AuthorDao) SelectById(ctx context.Context, id int64) (*Article, error) {
	article := &Article{}
	err := d.db.Find(article, id).Error
	return article, err
}

func (d *AuthorDao) Create(ctx context.Context, artEntity *Article) error {
	now := time.Now().UnixMilli()
	artEntity.Utime = now
	artEntity.Ctime = now
	return d.db.Create(artEntity).Error
}
