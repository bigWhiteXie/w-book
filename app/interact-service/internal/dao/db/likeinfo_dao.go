package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArticleStatusPublished = 1

	NoRowsAffected = errors.New("no rows affected")
)

type LikeInfo struct {
	Id     int64  `json:"" gorm:"primaryKey"`
	Biz    string `json:"" gorm:"type:varchar(122); uniqueIndex:biz_uid_idx"`
	BizId  int64  `json:"" gorm:"uniqueIndex:biz_uid_idx"`
	Uid    int64  `json:"" gorm:""`
	Status uint8  `json:"" gorm:""`
	Ctime  int64  `json:""`
	Utime  int64  `json:"" gorm:"uniqueIndex:biz_uid_idx`
}

func (a *LikeInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *LikeInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *LikeInfo) TableName() string {
	return "like_info"
}

type LikeInfoDao struct {
	// go get github.com/DATA-DOG/go-sqlmock
	db *gorm.DB
}

func NewLikeInfoDao(db *gorm.DB) *LikeInfoDao {
	return &LikeInfoDao{db: db}
}

func (d *LikeInfoDao) UpdateLikeInfo(ctx context.Context, uid int64, biz string, bizId int64, status uint8) error {
	now := time.Now().UnixMilli()
	likeInfo := &LikeInfo{
		Biz:    biz,
		BizId:  bizId,
		Uid:    uid,
		Status: uint8(status),
		Ctime:  now,
		Utime:  now,
	}
	// 尝试插入点赞记录，若冲突则更新status和utime
	res := d.db.Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"status": status,
				"utime":  now,
			}), // 更新字段
		},
	).Create(likeInfo)

	if res.Error != nil {
		return errors.Wrapf(res.Error, "[LikeInfoDao_UpdateLikeInfo] 插入点赞信息失败")
	}

	if res.RowsAffected == 0 {
		return NoRowsAffected
	}
	return nil
}

func (d *LikeInfoDao) FindLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) (*LikeInfo, error) {
	likeInfo := &LikeInfo{}
	res := d.db.Where("biz=? and biz_id=? and uid=?", biz, bizId, uid).First(likeInfo)
	if res.Error != nil {
		return nil, errors.Wrapf(res.Error, "[LikeInfoDao_FindLikeInfo] 查询点赞信息失败")
	}

	return likeInfo, nil
}

// 批量查询点赞状态
func (d *LikeInfoDao) BatchFindLikeInfo(ctx context.Context, uid int64, biz string, bizIds []int64) (map[int64]bool, error) {
	likeInfos := make([]*LikeInfo, 0, len(bizIds))

	// 使用IN查询批量获取点赞状态
	err := d.db.WithContext(ctx).
		Model(&LikeInfo{}).
		Select("biz_id, status").
		Where("uid = ? AND biz = ? AND biz_id IN (?)", uid, biz, bizIds).
		Find(&likeInfos).Error

	if err != nil {
		return nil, errors.Wrapf(err, "[LikeInfoDao_BatchFindLikeInfo] 批量查询失败")
	}

	// 初始化结果集，默认未点赞
	result := make(map[int64]bool, len(bizIds))
	for _, id := range bizIds {
		result[id] = false
	}

	// 标记已点赞的资源
	for _, like := range likeInfos {
		result[like.BizId] = like.Status == 1
	}

	return result, nil
}
