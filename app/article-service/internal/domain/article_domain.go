package domain

import "encoding/json"

type Article struct {
	Id      int64         `json:"id"`
	Title   string        `json:"title"`
	Content string        `json:"content"`
	Status  ArticleStatus `json:"status"`
	Author  Author        `json:"author"`
	Utime   int64         `json:"utime"`
	Ctime   int64         `json:"ctime"`
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
