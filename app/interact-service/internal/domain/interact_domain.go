package domain

import "encoding/json"

const (
	Like    = "like"
	Read    = "read"
	Collect = "collect"
)

type StatCnt struct {
	Biz        string `json:""`
	BizId      int64  `json:""`
	LikeCnt    int64  `json:""`
	ReadCnt    int64  `json:""`
	CollectCnt int64  `json:""`
}

func (a *StatCnt) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

func (a *StatCnt) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}
