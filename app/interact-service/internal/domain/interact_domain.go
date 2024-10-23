package domain

import "encoding/json"

const (
	Like    = "like"
	Read    = "read"
	Collect = "collect"
)

const (
	ReadEvtTopic   = "read-evt-topic"
	CreateEvtTopic = "create-evt-topic"
)

type Collection struct {
	Id    int64  `json:""`
	Name  string `json:""`
	Uid   int64  `json:""`
	Count int64  `json:""`
	Ctime int64  `json:""`
	Utime int64  `json:""`
}

type CollectionItem struct {
	Id     int64  `json:""`
	Uid    int64  `json:""`
	Cid    int64  `json:""`
	Biz    string `json:""`
	BizId  int64  `json:""`
	Ctime  int64  `json:""`
	Utime  int64  `json:""`
	Action uint8  // 0 取消收藏| 1 添加收藏
}

type StatCnt struct {
	Biz        string `json:""`
	BizId      int64  `json:""`
	LikeCnt    int64  `json:""`
	ReadCnt    int64  `json:""`
	CollectCnt int64  `json:""`
}

type ReadEvent struct {
	Biz   string `json:""`
	BizId int64  `json:""`
	Uid   int64  `json:""`
}

type Record struct {
	Biz   string `json:""` // 在唯一索引和查询索引中使用
	BizId int64  `json:""`
	UId   int64  `json:""`
}

// redis缓存需要实现该方法
func (a *StatCnt) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *StatCnt) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}
