// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.2

package types

import "codexie.com/w-book-interact/internal/domain"

type OpResourceReq struct {
	Biz        string `json:"biz"`
	BizId      int64  `json:"biz_id"`
	Action     uint8  `json:"action,optional"`
	Uid        int64   `json:"uid,optional"`
}

type CollectionReq struct {
	Id        int64 `json:"id,optional"`
	Name        string `json:"name"`
	Action      int  `json:"action,optional"`
	
}

type CollectResourceReq struct {
	Id        int `json:"id,optional"`
	Biz       string `json:"biz"`
	BizId     int64  `json:"biz_id"`
	Cid       int64  `json:"cid"`
	Action    int  `json:"action,optional"`
}

type TopLikeReq struct {
	Biz       string `path:"biz"`
}

func (req *CollectResourceReq) ToDomain(uid int64) *domain.CollectionItem {
	return &domain.CollectionItem{
		Id: int64(req.Id),
		Uid: uid,
		Biz: req.Biz,
		BizId: int64(req.BizId),
		Cid: int64(req.Cid),
		Action: uint8(req.Action),
	}
}

type EditArticleResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type ArticlePageReq struct {
	Page   int `form:"page"`
	Size   int `form:"size"`
}

type ArticleViewReq struct {
	Id    int64 `form:"id"`
	IsPublished bool `form:isPublished`
}
