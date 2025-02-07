package domain

import "codexie.com/w-book-account/api/pb"

// AccountType 账户类型
type AccountType uint8

func GetAccountType(tp pb.AccountType) AccountType {
	switch tp {
	case pb.AccountType_Reward:
		return AccountTypeReward
	default:
		return AccountTypeUnknown
	}
}

const (
	AccountTypeUnknown AccountType = 0
	AccountTypeReward  AccountType = 1
)

// Account 账户领域模型
type Account struct {
	Id          int64
	Uid         int64
	Account     int64
	AccountType AccountType
	Balance     int64
	Currency    string
}

type UpdateBalanceRequest struct {
	Uid         int64
	Account     int64
	AccountType AccountType
	Amount      int64
}

// AccountActivity 账户活动领域模型
type AccountActivity struct {
	Id          int64
	Uid         int64
	Account     int64
	AccountType AccountType
	Biz         string
	OutTradeNo  string
	Amount      int64
	Currency    string
}

// CreditItem 分账项
type CreditItem struct {
	Uid         int64
	Account     int64
	AccountType AccountType
	Amount      int64
	Currency    string
}

// CreditRequest 分账请求
type CreditRequest struct {
	Biz         string
	OutTradeNo  string
	CreditItems []CreditItem
}
