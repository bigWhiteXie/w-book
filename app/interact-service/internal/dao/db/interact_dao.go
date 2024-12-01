package db

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Interaction struct {
	Id         int64  `json:"" gorm:"primaryKey,autoIncrement"`
	Biz        string `json:"" gorm:"type:varchar(128); uniqueIndex:biz_type_id"`
	BizId      int64  `json:"" gorm:"uniqueIndex:biz_type_id"`
	LikeCnt    int64  `json:""`
	ReadCnt    int64  `json:""`
	CollectCnt int64  `json:""`

	Ctime int64 `json:""`
	Utime int64 `json:""`
}

type InteractDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewInteractDao(db *gorm.DB) *InteractDao {
	return &InteractDao{db: db}
}

func (d *InteractDao) GetTopResourcesByLikes(resourceType string, limit int) ([]Interaction, error) {
	var resources []Interaction
	if err := d.db.Where("biz = ?", resourceType).Order("like_cnt DESC").Limit(limit).Find(&resources).Error; err != nil {
		return nil, errors.Wrapf(err, "[InteractDao_GetTopResourcesByLikes] 查询点赞数失败,limit:%d", limit)
	}

	return resources, nil
}

func (d *InteractDao) CrateCntData(ctx context.Context, biz string, bizId int64) (*Interaction, error) {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:     biz,
		BizId:   bizId,
		LikeCnt: 0,
		Utime:   now,
		Ctime:   now,
	}
	if err := d.db.Create(interaction).Error; err != nil {
		return nil, errors.Wrapf(err, "[InteractDao_CrateCntData] 创建点赞数失败,biz:%s,bizId:%d", biz, bizId)
	}

	return interaction, nil
}
func (d *InteractDao) FindInteractByBiz(ctx context.Context, biz string, bizId int64) (*Interaction, error) {
	model := &Interaction{}
	result := d.db.Where("biz=? and biz_id=?", biz, bizId).First(model)
	if result.Error != nil {
		return nil, errors.Wrapf(result.Error, "[InteractDao_FindInteractByBiz] 查询点赞数失败,biz:%s,bizId:%d", biz, bizId)
	}

	return model, nil
}

func (d *InteractDao) IncreLike(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:     biz,
		BizId:   bizId,
		LikeCnt: 1,
		Utime:   now,
		Ctime:   now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("`like_cnt`+1"),
				"utime":    now,
			}), // 更新字段
		},
	).Create(interaction)
	if result.Error != nil {
		return errors.Wrapf(result.Error, "[InteractDao_DecreLike] 更新点赞数失败,biz:%s,bizId:%d", biz, bizId)
	}

	return nil
}

func (d *InteractDao) DecreLike(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Model(&Interaction{}).Where("biz=? and biz_id=?", biz, bizId).UpdateColumns(map[string]interface{}{
		"like_cnt": gorm.Expr("like_cnt - ?", 1),
		"utime":    now,
	})

	if result.Error != nil {
		return errors.Wrapf(result.Error, "[InteractDao_DecreLike] 更新点赞数失败,biz:%s,bizId:%d", biz, bizId)
	}

	if result.RowsAffected == 0 {
		return errors.Wrap(fmt.Errorf("[InteractDao_DecreLike] 用户[%d]点赞不存在的数据", ctx.Value("id")), "")
	}
	return nil
}

func (d *InteractDao) IncreCollectCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:        biz,
		BizId:      bizId,
		CollectCnt: 1,
		Utime:      now,
		Ctime:      now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"utime":       now,
			}), // 更新字段
		},
	).Create(interaction)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) DecreCollectCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).Updates(map[string]any{
		"collect_cnt": gorm.Expr("`collect_cnt` - 1"),
		"utime":       now,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) IncreRead(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	interaction := &Interaction{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Utime:   now,
		Ctime:   now,
	}
	result := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"read_cnt": gorm.Expr("`read_cnt`+ 1"),
				"utime":    now,
			}), // 更新字段
		},
	).Create(interaction)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) BatchIncreRead(ctx context.Context, bizs []string, bizIds []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewInteractDao(tx)
		for i := 0; i < len(bizs); i++ {
			err := dao.IncreRead(ctx, bizs[i], bizIds[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *InteractDao) DecreReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()

	result := d.db.Where("biz=? and biz_id=?", biz, bizId).Updates(map[string]any{
		"read_cnt": gorm.Expr("`read_cnt` - 1"),
		"utime":    now,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *InteractDao) GetInteractions(ctx context.Context, biz string, bizIds []int64) ([]Interaction, error) {
	var interactions []Interaction
	err := d.db.WithContext(ctx).
		Where("biz = ? AND biz_id IN (?)", biz, bizIds).
		Find(&interactions).Error

	if err != nil {
		return nil, errors.Wrap(err, "[InteractDao_FindInteractionsByBiz] 数据库查询失败")
	}

	idToInteraction := make(map[int64]Interaction, len(interactions))
	for _, interaction := range interactions {
		idToInteraction[interaction.BizId] = interaction
	}

	result := make([]Interaction, 0, len(bizIds))
	for _, id := range bizIds {
		if interaction, exists := idToInteraction[id]; exists {
			result = append(result, interaction)
		}
	}

	return result, nil
}
