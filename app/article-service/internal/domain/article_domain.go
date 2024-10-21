package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Status  ArticleStatus
	Author  Author
	Utime   int64
	Ctime   int64
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
	Id   int64
	Name string
}
