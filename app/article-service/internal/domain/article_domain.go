package domain

import "encoding/json"

const (
	Biz                = "article"
	ReadTopic          = "read-evt-topic"
	ArticleCreateTopic = "create-evt-topic"
)

type ReadEvent struct {
	Biz   string `json:""`
	BizId int64  `json:""`
	Uid   int64  `json:""`
}

type Article struct {
	StatInfo

	Id      int64         `json:"id"`
	Title   string        `json:"title"`
	Content string        `json:"content"`
	Status  ArticleStatus `json:"status"`
	Author  Author        `json:"author"`
	Utime   int64         `json:"utime"`
	Ctime   int64         `json:"ctime"`
}

type StatInfo struct {
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
}

func (a *Article) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Article) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

type ArticleArray []*Article

// 序列化
func (m ArticleArray) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

// 反序列化
func (m ArticleArray) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)

}

type ArticleStatus uint8

func ArticleStatusFromUint8(n uint8) ArticleStatus {
	switch n {
	case 1:
		return ArticleUnpublishedStatus
	case 2:
		return ArticlePublishedStatus
	case 3:
		return ArticleWithdrawStatus
	default:
		return ArticleUnknowStatus
	}
}
func (s ArticleStatus) ToUnit8() uint8 {
	return uint8(s)
}

const (
	ArticleUnknowStatus ArticleStatus = iota
	ArticleUnpublishedStatus
	ArticlePublishedStatus
	ArticleWithdrawStatus
)

type Author struct {
	Id   int64  `json:"author_id"`
	Name string `json:"name"`
}
