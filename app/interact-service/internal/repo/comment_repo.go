package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/repo"
	"codexie.com/w-book-interact/internal/dao/cache"
	"codexie.com/w-book-interact/internal/dao/db"
	"codexie.com/w-book-interact/internal/domain"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type ICommentRepo interface {
	GetRootComments(ctx context.Context, biz string, bizID int64, offset, size int, lastScore int64) ([]*domain.Comment, error)
	GetChildComments(ctx context.Context, rootID int64, lastCTime int64, offset, size int) ([]*domain.Comment, error)
	CreateComment(ctx context.Context, biz string, bizID int64, parentID int64, content string, uid int64) (*domain.Comment, error)
	LikeComment(ctx context.Context, id int64, uid int64, isLike bool) error
	DeleteComment(ctx context.Context, id int64) error
}

type commentRepo struct {
	*repo.BaseRepo
	cache       cache.CommentCache
	likeInfoDao *db.LikeInfoDao
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
		return nil, fmt.Errorf("查询根评论失败: %w", err)
	}
	// 转换为domain对象
	wg := errgroup.Group{}
	rootDomains := make([]*domain.Comment, 0, len(dbComments))
	for _, c := range dbComments {
		rootDomain := db.CommentDo2Domain(ctx, c)

		wg.Go(func() error {
			childComments, err := commentDao.GetChildCommentsByRootIDs(ctx, rootDomain.ID, 3)
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
		return nil, fmt.Errorf("查询子评论失败: %w", err)
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
				return nil, codeerr.LogCodeError(ctx, "父评论不存在", err, "父评论不存在")
			}
			return nil, codeerr.LogCodeError(ctx, "数据库异常", err, "获取父评论异常导致发表评论失败,comment:%v", comment)
		}
		comment.RootID = parentComment.RootID
	}

	if err := commentDao.CreateCommentV2(ctx, comment); err != nil {
		return nil, codeerr.LogCodeError(ctx, "数据库异常", err, "创建评论失败,comment:%v", comment)
	}

	return db.CommentDo2Domain(ctx, comment), nil
}

func (r *commentRepo) LikeComment(ctx context.Context, id int64, uid int64, isLike bool) error {
	//0.开启事务
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		likeInfoDao := db.NewLikeInfoDao(tx)
		if isLike {
			if err := likeInfoDao.Like(ctx, uid, "comment", id); err != nil {
				if err == db.NoRowsAffected {
					return nil
				}
				return codeerr.LogCodeError(ctx, "数据库异常", err, "点赞失败,comment_id:%d,uid:%d", id, uid)
			}
		} else {
			if err := likeInfoDao.UnLike(ctx, uid, "comment", id); err != nil {
				if err == db.NoRowsAffected {
					return nil
				}
				return codeerr.LogCodeError(ctx, "数据库异常", err, "取消点赞失败,comment_id:%d,uid:%d", id, uid)
			}
		}
		commentDao := db.NewCommentDAO(tx)
		return commentDao.LikeComment(ctx, id, isLike)
	})
}

func (r *commentRepo) DeleteComment(ctx context.Context, id int64) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		commentDao := db.NewCommentDAO(tx)
		return commentDao.DeleteComment(ctx, id)
	})
}
