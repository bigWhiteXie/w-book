package domain

type RewardRecord struct {
	Id         int64  `json:""`
	Uid        int64  `json:""`
	Biz        string `json:""`
	Platform   string `json:""`
	Amt        int64  `json:""`
	ResourceId int64  `json:""`
	OutTradeNo string `json:""`
	Status     string `json:""`
}
