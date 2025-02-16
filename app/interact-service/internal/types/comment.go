package types

type AddCommentReq struct {
	Biz      string `json:"biz"`       // 业务类型
	BizID    int64  `json:"biz_id"`    // 业务ID
	ParentID int64  `json:"parent_id"` // 父评论ID
	Content  string `json:"content"`   // 评论内容
}

type GetRootCommentsReq struct {
	Biz       string `json:"biz"`        // 业务类型
	BizID     int64  `json:"biz_id"`     // 业务ID
	Offset    int    `json:"offset"`     // 分页偏移
	Size      int    `json:"size"`       // 分页大小
	LastScore int64  `json:"last_score"` // 最后一条评论的分数（用于分页）
}

type GetChildCommentsReq struct {
	RootID    int64 `json:"root_id"`    // 根评论ID
	LastCTime int64 `json:"last_ctime"` // 最后一条评论的创建时间（用于分页）
	Offset    int   `json:"offset"`     // 分页偏移
	Size      int   `json:"size"`       // 分页大小
}

type LikeCommentReq struct {
	CommentID int64 `json:"comment_id"` // 评论ID
	IsLike    bool  `json:"is_like"`    // 是否点赞
}
