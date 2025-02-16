package domain

import "time"

const (
	RootCommetCacheSize       = 20
	ChildCommetCacheSize      = 30
	DefaultChildCommentsLimit = 3

	CommentBiz = "comment"

	HotArticleIdsKey = "hot:article:ids"

	CommentEvtTopic = "comment_events"

	CreateCommentEvt = "create_comment"
	LikeCommentEvt   = "like_comment"
)

type Comment struct {
	ID       int64      `json:"id"`
	Uid      int64      `json:"uid"`
	Biz      string     `json:"biz"`
	BizID    int64      `json:"biz_id"`
	ChildNum int64      `json:"child_num"`
	ParentID int64      `json:"parent_id"`
	LikeCnt  int        `json:"like_cnt"`
	Content  string     `json:"content"`
	Childs   []*Comment `json:"childs"`
	Ctime    time.Time  `json:"created_at"`
	Score    int64      `json:"score"`
	RootID   int64      `json:"root_id"`
	IsLike   bool       `json:"is_like"`
}

type CommentEvent struct {
	CommentID int64  `json:"comment_id"`
	RootID    int64  `json:"root_id"` // 根评论ID（如果是根评论则为0）
	Action    string `json:"action"`  // 事件类型
}
