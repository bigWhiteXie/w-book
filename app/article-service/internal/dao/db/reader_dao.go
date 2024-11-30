package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArticleStatusPublished = 1
)

type PublishedArticle struct {
	Id       int64  `json:"",gorm:"primaryKey"`
	Title    string `json:""`
	Content  string `json:"",gorm:"type:blob"`
	AuthorId int64  `json:"",gorm:"index:idx_uid_uptime"`
	Status   uint8  `json:"",gorm:""`
	Ctime    int64  `json:"",`
	Utime    int64  `json:"",gorm:"index:idx_uid_uptime"`
}

func (a *PublishedArticle) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *PublishedArticle) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
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
		return errors.Wrapf(res.Error, "[ReaderDao_UpdateById] 更新文章失败,id=%d", artEntity.Id)
	}
	if res.RowsAffected == 0 {
		return errors.Wrap(fmt.Errorf("[ReaderDao_UpdateById] 用户[%d]修改他人文章[%s]", artEntity.AuthorId, artEntity.Id), "")
	}

	return nil
}

func (d *ReaderDao) Create(ctx context.Context, artEntity *PublishedArticle) error {
	now := time.Now().UnixMilli()
	artEntity.Ctime = now
	artEntity.Utime = now
	if err := d.db.Create(artEntity).Error; err != nil {
		return errors.Wrapf(err, "[ReaderDao_Create] 创建文章失败,id=%d", artEntity.Id)
	}

	return nil
}

func (d *ReaderDao) Save(ctx context.Context, artEntity *PublishedArticle) error {
	if err := d.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},                                   // 以 ID 作为冲突判断
			DoUpdates: clause.AssignmentColumns([]string{"title", "content", "utime"}), // 更新字段
		},
	).Create(&artEntity).Error; err != nil {
		return errors.Wrapf(err, "[ReaderDao_Save] 保存文章失败,id=%d", artEntity.Id)
	}

	return nil
}

func (d *ReaderDao) FindById(ctx context.Context, id int64) (*PublishedArticle, error) {
	art := &PublishedArticle{}
	err := d.db.First(art, "id=?", id).Error
	if err != nil {
		return art, errors.Wrapf(err, "[ReaderDao_FindById] 查询文章失败,id=%d", id)
	}
	return art, nil
}

func (d *ReaderDao) FindShortArticlesBatch(ids []int64) ([]*PublishedArticle, error) {
	arts := make([]*PublishedArticle, 0, len(ids))
	err := d.db.Find(&arts, "id in (?)", ids).Error
	if err != nil {
		return nil, errors.Wrapf(err, "[ReaderDao_FindShortArticlesBatch] 查询文章失败,ids=%v", ids)
	}
	// 构建 ID 到文章的映射
	artMap := make(map[int64]*PublishedArticle, len(arts))
	for _, art := range arts {
		artMap[art.Id] = art
	}

	// 按照 ids 顺序构建结果
	result := make([]*PublishedArticle, 0, len(ids))
	for _, id := range ids {
		if art, exists := artMap[id]; exists {
			result = append(result, art)
		}
	}

	return result, nil
}
