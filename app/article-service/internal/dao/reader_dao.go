package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArticleStatusPublished = 1
)

type PublishedArticle struct {
	Id       int64 `gorm:"primaryKey"`
	Title    string
	Content  string `gorm:"type:blob"`
	AuthorId int64  `gorm:"index:idx_uid_uptime"`
	Status   uint8  `gorm:""`
	Ctime    int64
	Utime    int64 `gorm:"index:idx_uid_uptime"`
}

type ReaderDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewReaderDao(db *gorm.DB) *ReaderDao {
	return &ReaderDao{db: db}
}

func (d *ReaderDao) UpdateById(ctx context.Context, artEntity *PublishedArticle) error {
	res := d.db.Model(&PublishedArticle{}).Where(
		"id=? and author_id=?", artEntity.Id, artEntity.AuthorId,
	).Updates(map[string]interface{}{
		"title":   artEntity.Title,
		"content": artEntity.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  ArticleStatusPublished,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New(fmt.Sprintf("用户[%d] 修改非自己的发布文章%d", artEntity.AuthorId, artEntity.Id))
	}
	return nil
}

func (d *ReaderDao) Create(ctx context.Context, artEntity *PublishedArticle) error {
	now := time.Now().UnixMilli()
	artEntity.Ctime = now
	artEntity.Utime = now
	return d.db.Create(artEntity).Error
}

func (d *ReaderDao) Save(ctx context.Context, artEntity *PublishedArticle) error {
	err := d.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},                                   // 以 ID 作为冲突判断
			DoUpdates: clause.AssignmentColumns([]string{"title", "content", "utime"}), // 更新字段
		},
	).Where("author_id=?", artEntity.AuthorId).Create(&artEntity).Error
	return err
}

func (d *ReaderDao) FindById(ctx context.Context, id int64) (*PublishedArticle, error) {
	art := &PublishedArticle{}
	err := d.db.First(art, "id=?", id).Error
	if err != nil {
		return art, err
	}
	return art, nil
}
