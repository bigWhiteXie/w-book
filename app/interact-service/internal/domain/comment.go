package domain

import "time"

const (
	RootCommetCacheSize  = 20
	ChildCommetCacheSize = 30
)

type Comment struct {
	ID       int64      `json:"id"`
	Uid      int64      `json:"uid"`
	ChildNum int64      `json:"child_num"`
	LikeCnt  int        `json:"like_cnt"`
	Content  string     `json:"content"`
	Childs   []*Comment `json:"childs"`
	Ctime    time.Time  `json:"created_at"`
	Score    int64      `json:"score"`
	RootID   int64      `json:"root_id"`
}
