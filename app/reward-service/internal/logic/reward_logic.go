package logic

import (
	"context"
	"strconv"

	"codexie.com/w-book-common/kafka/producer"
	"codexie.com/w-book-payment/api/pb"
	"codexie.com/w-book-reward/internal/domain"
	"codexie.com/w-book-reward/internal/repo"
	"github.com/google/uuid"
)

var (
	articleBiz = "article"
)

type RewardLogic struct {
	rewardRepo repo.IRewardRepository
	paymentRpc pb.PaymentClient
	producer   producer.Producer
}

func NewRewardLogic(rewardRepo *repo.RewardRepository, paymentClient pb.PaymentClient) *RewardLogic {
	return &RewardLogic{rewardRepo: rewardRepo, paymentRpc: paymentClient}
}

func (l *RewardLogic) RewardPcWEB(ctx context.Context, reward *domain.RewardRecord) (string, error) {
	//生成未支付的打赏记录
	outTradeNo := uuid.New().String()
	reward.OutTradeNo = outTradeNo
	id, err := l.rewardRepo.CreateReward(ctx, reward)
	if err != nil {
		return "", err
	}
	//调用支付服务，返回支付链接
	resp, err := l.paymentRpc.PrePay(ctx, &pb.PrepayReq{
		Biz:         "reward",
		OutTradeNo:  outTradeNo,
		Subject:     "打赏" + reward.Biz,
		Platform:    reward.Platform,
		TotalAmount: strconv.Itoa(int(reward.Amt)),
	})
	if err != nil {
		l.rewardRepo.UpdateReward(ctx, &domain.RewardRecord{Id: id, Status: "prepay_fail"})
		return "", err
	}

	return resp.PayUrl, err
}
