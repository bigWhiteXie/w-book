package repo

import (
	"context"
	"database/sql"
	"time"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type ICommentRepo interface {
	GetRootComments(ctx context.Context, biz string, bizID int64, offset, size int, lastScore int64) ([]*domain.Comment, error)
	GetChildComments(ctx context.Context, rootID int64, lastCTime int64, offset, size int) ([]*domain.Comment, error)
	CreateComment(ctx context.Context, biz string, bizID int64, parentID int64, content string, uid int64) (*domain.Comment, error)
	LikeComment(ctx context.Context, id int64, uid int64, isLike bool) error
	DeleteComment(ctx context.Context, id int64, uid int64) error
	GetCommentByID(ctx context.Context, id int64) (*domain.Comment, error)
	GetRootCommentsByBizIDs(ctx context.Context, biz string, bizIDs []int64, size int) ([]*domain.Comment, error)
	HandleCommentCreateEvent(ctx context.Context, commentID, rootID int64) error
	HandleLikeCommentEvent(ctx context.Context, commentID int64, isLike bool, uid int64) error
	GetLikeStatusByUid(ctx context.Context, uid int64, commentIDs []int64) (map[int64]bool, error)
	GetCommentLikeUserIDs(ctx context.Context, commentID int64) ([]int64, error)
}

type commentRepo struct {
	*repo.BaseRepo
	cache cache.CommentCache
}

func NewCommentRepo(gormDb *gorm.DB, cache cache.CommentCache) ICommentRepo {
	return &commentRepo{BaseRepo: repo.NewBaseRepo(gormDb), cache: cache}
}

func (r *commentRepo) GetRootComments(ctx context.Context, biz string, bizID int64, offset, size int, lastScore int64) ([]*domain.Comment, error) {
	// 检查是否使用缓存
	if offset+size < 20 {
		// 直接调用CommentCache接口获取缓存
		comments, err := r.cache.GetRootComments(ctx, biz, bizID, offset, size)
		if err == nil && len(comments) > 0 {
			return comments, nil
		}
	}

	// 从数据库查询
	commentDao := db.NewCommentDAO(r.GetDB())
	dbComments, err := commentDao.GetRootCommentsByScore(ctx, biz, bizID, lastScore, size)
	if err != nil {
		return nil, err
	}
	// 转换为domain对象
	wg := errgroup.Group{}
	rootDomains := make([]*domain.Comment, 0, len(dbComments))
	for _, c := range dbComments {
		rootDomain := db.CommentDo2Domain(ctx, c)

		wg.Go(func() error {
			childComments, err := commentDao.GetChildCommentsByRootIDs(ctx, rootDomain.ID, domain.DefaultChildCommentsLimit)
			if err != nil {
				return err
			}
			childDomains := make([]*domain.Comment, 0, len(childComments))
			for _, c := range childComments {
				childDomains = append(childDomains, db.CommentDo2Domain(ctx, c))
			}
			rootDomain.Childs = childDomains
			return nil
		})

		rootDomains = append(rootDomains, rootDomain)
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return rootDomains, nil
}

func (r *commentRepo) GetChildComments(ctx context.Context, rootID int64, lastCTime int64, offset, size int) ([]*domain.Comment, error) {
	// 检查是否使用缓存
	if offset+size <= 30 {
		// 直接调用CommentCache接口获取缓存
		comments, err := r.cache.GetChildComments(ctx, rootID, offset, size)
		if err == nil && len(comments) > 0 {
			return comments, nil
		}
	}

	// 从数据库查询
	commentDao := db.NewCommentDAO(r.GetDB())
	dbComments, err := commentDao.GetChildCommentsByTime(ctx, rootID, lastCTime, size)
	if err != nil {
		return nil, err
	}

	// 转换为domain对象
	var result []*domain.Comment
	for _, c := range dbComments {
		domainComment := db.CommentDo2Domain(ctx, c)
		result = append(result, domainComment)
	}

	return result, nil
}

func (r *commentRepo) CreateComment(ctx context.Context, biz string, bizID int64, parentID int64, content string, uid int64) (*domain.Comment, error) {
	now := time.Now().UnixMilli()
	comment := &db.Comment{
		Biz:     biz,
		BizID:   bizID,
		Content: content,
		Uid:     uid,
		Ctime:   now,
	}

	if parentID > 0 {
		comment.ParentID = sql.NullInt64{Int64: parentID, Valid: true}
	}

	commentDao := db.NewCommentDAO(r.GetDB())

	// 获取parent评论
	if parentID > 0 {
		parentComment, err := commentDao.GetCommentByID(ctx, parentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, codeerr.LogCodeError(ctx, "父评论不存在", "父评论不存在")
			}
			return nil, codeerr.LogCodeError(ctx, "数据库异常", "获取父评论异常导致发表评论失败,comment:%v", comment)
		}
		comment.RootID = parentComment.RootID
	}

	if err := commentDao.CreateCommentV2(ctx, comment); err != nil {
		return nil, codeerr.LogCodeError(ctx, "数据库异常", "创建评论失败,comment:%v", comment)
	}

	return db.CommentDo2Domain(ctx, comment), nil
}

func (r *commentRepo) LikeComment(ctx context.Context, id int64, uid int64, isLike bool) error {
	if isLike {
		// 设置点赞状态缓存为true
		if err := r.cache.SetLikesStatus(ctx, uid, []int64{id}, map[int64]bool{id: true}); err != nil {
			return codeerr.LogCodeError(ctx, "redis异常", "设置点赞状态缓存失败 uid:%d commentID:%d", uid, id)
		}
	} else {
		// 设置点赞状态缓存为false
		if err := r.cache.SetLikesStatus(ctx, uid, []int64{id}, map[int64]bool{id: false}); err != nil {
			return codeerr.LogCodeError(ctx, "redis异常", "设置点赞状态缓存失败 uid:%d commentID:%d", uid, id)
		}
	}
	commentDao := db.NewCommentDAO(r.GetDB())

	return commentDao.LikeComment(ctx, id, isLike)
}

func (r *commentRepo) DeleteComment(ctx context.Context, id int64, uid int64) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		commentDao := db.NewCommentDAO(tx)
		return commentDao.DeleteComment(ctx, id, uid)
	})
}

func (r *commentRepo) GetCommentByID(ctx context.Context, id int64) (*domain.Comment, error) {
	comment, err := db.NewCommentDAO(r.GetDB()).GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, codeerr.LogCodeError(ctx, "评论不存在", "查询评论失败 id:%d", id)
		}
		return nil, codeerr.LogCodeError(ctx, "数据库异常", "查询评论失败 id:%d", id)
	}
	return db.CommentDo2Domain(ctx, comment), nil
}

func (r *commentRepo) GetRootCommentsByBizIDs(ctx context.Context, biz string, bizIDs []int64, size int) ([]*domain.Comment, error) {
	commentDao := db.NewCommentDAO(r.GetDB())

	// 从数据库查询这些资源的前n条根评论
	dbComments, err := commentDao.GetRootCommentsByBizIDs(ctx, biz, bizIDs, size)
	if err != nil {
		return nil, codeerr.LogCodeError(ctx, "查询多资源评论失败", "biz:%s bizIDs:%v", biz, bizIDs)
	}

	// 使用errgroup并行获取子评论
	var (
		wg       = errgroup.Group{}
		comments = make([]*domain.Comment, 0, len(dbComments))
	)

	for _, c := range dbComments {
		comment := db.CommentDo2Domain(ctx, c)
		comments = append(comments, comment)

		wg.Go(func() error {
			childComments, err := commentDao.GetChildCommentsByRootIDs(
				ctx,
				comment.ID,
				domain.DefaultChildCommentsLimit,
			)
			if err != nil {
				return codeerr.LogCodeError(ctx, "获取子评论失败",
					"rootID:%d", comment.ID)
			}

			childDomains := make([]*domain.Comment, 0, len(childComments))
			for _, cc := range childComments {
				childDomains = append(childDomains, db.CommentDo2Domain(ctx, cc))
			}
			comment.Childs = childDomains
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return comments, nil
}

// 处理评论事件（状态更新和计数调整）
func (r *commentRepo) HandleCommentCreateEvent(ctx context.Context, commentID, rootID int64) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		// 1. 更新评论状态
		if err := tx.Model(&db.Comment{}).
			Where("id = ?", commentID).
			Update("status", 1).Error; err != nil {
			return codeerr.LogCodeError(ctx, "更新评论状态失败", "更新评论状态失败 commentID:%d", commentID)
		}

		// 2. 更新根评论子评论数
		if rootID > 0 && commentID != rootID {
			if err := tx.Model(&db.Comment{}).
				Where("id = ?", rootID).
				Update("child_num", gorm.Expr("child_num + 1")).Error; err != nil {
				return codeerr.LogCodeError(ctx, "更新根评论子评论数失败", "更新根评论子评论数失败 rootID:%d", rootID)
			}
		}
		commentDao := db.NewCommentDAO(tx)
		// 重新计算根评论分数
		if err := commentDao.RefreshCommentScore(ctx, rootID); err != nil {
			return codeerr.LogCodeError(ctx, "刷新评论分数失败", "刷新评论分数失败 rootID:%d", rootID)
		}
		return nil
	})
}

// 处理点赞事件（刷新评论分数）
func (r *commentRepo) HandleLikeCommentEvent(ctx context.Context, commentID int64, isLike bool, uid int64) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		likeInfoDao := db.NewLikeInfoDao(tx)
		if isLike {
			if err := likeInfoDao.Like(ctx, uid, "comment", commentID); err != nil {
				if err == db.NoRowsAffected {
					return nil
				}
				return codeerr.LogCodeError(ctx, "数据库异常", "点赞失败,comment_id:%d,uid:%d", commentID, uid)
			}
		} else {
			if err := likeInfoDao.UnLike(ctx, uid, "comment", commentID); err != nil {
				if err == db.NoRowsAffected {
					return nil
				}
				return codeerr.LogCodeError(ctx, "数据库异常", "取消点赞失败,comment_id:%d,uid:%d", commentID, uid)
			}
		}

		commentDao := db.NewCommentDAO(tx)
		if err := commentDao.LikeComment(ctx, commentID, isLike); err != nil {
			return codeerr.LogCodeError(ctx, "点赞失败", "点赞失败 commentID:%d", commentID)
		}

		comment, err := r.GetCommentByID(ctx, commentID)
		if err != nil {
			return codeerr.LogCodeError(ctx, "数据库异常", "获取评论异常导致无法判断是否应该刷新分数 commentID:%d", commentID)
		}
		if comment.RootID != commentID {
			return nil
		}

		return commentDao.RefreshCommentScore(ctx, comment.RootID)
	})
}

// 查询该评论的点赞用户id
func (r *commentRepo) GetCommentLikeUserIDs(ctx context.Context, commentID int64) ([]int64, error) {
	likeInfoDao := db.NewLikeInfoDao(r.GetDB())
	return likeInfoDao.GetLikeStatus(ctx, domain.CommentBiz, commentID)
}

// 查询该用户在关于批评论的点赞状态
func (r *commentRepo) GetLikeStatusByUid(ctx context.Context, uid int64, commentIDs []int64) (map[int64]bool, error) {
	// 1. 从缓存获取已存在的点赞状态
	cachedStatus, err := r.cache.GetLikesStatus(ctx, uid, commentIDs)
	if err != nil {
		return nil, codeerr.LogCodeError(ctx, "获取点赞缓存失败", "uid:%d commentIDs:%v", uid, commentIDs)
	}

	// 2. 找出未命中的评论ID
	var missingIDs []int64
	for _, id := range commentIDs {
		if _, exists := cachedStatus[id]; !exists {
			missingIDs = append(missingIDs, id)
		}
	}

	if len(missingIDs) > 0 {
		likeInfoDao := db.NewLikeInfoDao(r.GetDB())
		// 3. 从数据库查询未命中的点赞状态
		dbStatus, err := likeInfoDao.BatchFindLikeInfo(ctx, uid, domain.CommentBiz, missingIDs)
		if err != nil {
			return nil, err
		}

		// 4. 合并结果并更新缓存
		for id, status := range dbStatus {
			cachedStatus[id] = status
		}
		if err := r.cache.SetLikesStatus(ctx, uid, missingIDs, dbStatus); err != nil {
			logx.WithContext(ctx).Errorf("更新点赞缓存失败: %v", err)
		}
	}

	return cachedStatus, nil
}
