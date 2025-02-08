package logic

import (
	"context"
	"strconv"
	"strings"

	account_pb "codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-payment/api/pb"

	"codexie.com/w-book-payment/pkg/constant"
	"codexie.com/w-book-reward/internal/dao/db"
	"codexie.com/w-book-reward/internal/domain"
	"codexie.com/w-book-reward/internal/repo"
	"codexie.com/w-book-reward/internal/types"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	articleBiz = "article"
)

type RewardLogic struct {
	rewardRepo    repo.IRewardRepository
	paymentRpc    pb.PaymentClient
	accountClient account_pb.AccountClient
}

func NewRewardLogic(rewardRepo *repo.RewardRepository, paymentClient pb.PaymentClient, accountClient account_pb.AccountClient) *RewardLogic {
	return &RewardLogic{rewardRepo: rewardRepo, paymentRpc: paymentClient, accountClient: accountClient}
}

func (l *RewardLogic) RewardPcWEB(ctx context.Context, reward *domain.RewardRecord) (*types.RewardResp, error) {
	//生成未支付的打赏记录
	outTradeNo := uuid.New().String()
	reward.OutTradeNo = outTradeNo
	id, err := l.rewardRepo.CreateReward(ctx, reward)
	if err != nil {
		return nil, err
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
		l.rewardRepo.UpdateReward(ctx, &domain.RewardRecord{Id: id, Status: constant.FailPayStatus})
		return nil, err
	}

	return &types.RewardResp{
		PayUrl:     resp.PayUrl,
		OutTradeNo: outTradeNo,
	}, nil
}

func (l *RewardLogic) RewardCredit(ctx context.Context, outTradeNo, status string, amt int64) (err error) {
	bizParams := strings.SplitN(outTradeNo, "-", 2)
	outTradeNo = bizParams[1]
	reward, err := l.rewardRepo.GetRewardByOutTradeNo(context.Background(), outTradeNo)
	if err != nil {
		return err
	}

	if err := l.rewardRepo.UpStatusByOutTradeNo(context.Background(), outTradeNo, status); err != nil {
		logx.Errorw("修改打赏记录支付状态失败", logx.LogField{Key: "outTradeNo", Value: outTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
		if err == db.NoRowsAffectedErr {
			return nil
		}
		return err
	}

	userAmt := int64(float64(reward.Amt) * 0.9)
	// 支付成功成功，进行分账
	if status == constant.SuceessPayStatus {
		if _, err := l.accountClient.Credit(context.Background(), &account_pb.CreditRequest{
			Biz:        "reward",
			OutTradeNo: outTradeNo,
			CreditItems: []*account_pb.CreditItem{
				&account_pb.CreditItem{
					Uid:         reward.AuthorId,
					Amt:         userAmt,
					AccountType: account_pb.AccountType_Reward,
					Currency:    constant.CNY,
				},
				{
					Amt:         reward.Amt - userAmt,
					AccountType: account_pb.AccountType_Reward,
					Currency:    constant.CNY,
				},
			},
		}); err != nil {
			logx.Errorw("打赏分账异常",
				logx.LogField{Key: "detail", Value: "调用分账服务grpc的Credit接口失败"},
				logx.LogField{Key: "outTradeNo", Value: outTradeNo},
				logx.LogField{Key: "cause", Value: err.Error()},
				logx.LogField{Key: "uid", Value: reward.AuthorId},
				logx.LogField{Key: "status", Value: reward.Status},
			)
		}
	}

	return nil
}
