package event

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-reward/internal/domain"
	"codexie.com/w-book-reward/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	rewardTopic = "reward-payment-callback"
)

type PayCallbackEvtListener struct {
	accountClient pb.AccountClient
	rewardRepo    repo.IRewardRepository
	batchConsumer *consumer.BatchConsumer[domain.PayMessage]
}

func NewBatchReadEventListener(client sarama.Client, accountClient pb.AccountClient) *PayCallbackEvtListener {
	l := &PayCallbackEvtListener{
		accountClient: accountClient,
	}
	l.batchConsumer = consumer.NewBatchConsumer[domain.PayMessage](rewardTopic, client, "reward-service-group", 100, l.handleBatchPayMsg)

	return l
}

func (c *PayCallbackEvtListener) handleBatchPayMsg(eventBatch []domain.PayMessage, msgs []*sarama.ConsumerMessage) error {
	logx.WithContext(context.Background()).Infof("批量阅读事件,长度:%d", len(eventBatch))
	for _, msg := range eventBatch {
		reward, err := c.rewardRepo.GetRewardByOutTradeNo(context.Background(), msg.OutTradeNo)
		if err != nil {
			logx.Errorw("获取打赏记录失败", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
			continue
		}
		if err := c.rewardRepo.UpdateReward(context.Background(), &domain.RewardRecord{
			Id:     reward.Id,
			Status: msg.Status,
		}); err != nil {
			logx.Errorw("更新打赏记录失败", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
			// todo 触发告警
		}

		if reward.Status == "success" {
			// 打赏成功，给作者增加收益
			if _, err := c.accountClient.Credit(context.Background(), &pb.CreditRequest{
				Biz:        "reward",
				OutTradeNo: reward.OutTradeNo,
				CreditItems: []*pb.CreditItem{
					&pb.CreditItem{
						Uid:         reward.AuthorId,
						Amt:         reward.Amt * 9 / 10,
						AccountType: pb.AccountType_Reward,
						Currency:    "CNY",
					},
					{
						Amt:         reward.Amt - reward.Amt*9/10,
						AccountType: pb.AccountType_Reward,
						Currency:    "CNY",
					},
				},
			}); err != nil {
				logx.Errorw("给作者增加收益失败", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
				// todo 触发告警
			}
		}
	}
	//打赏金额平台分成0.1，用户分成0.9
	return nil
}

func (c *PayCallbackEvtListener) Start() {
	c.batchConsumer.StartListner()
}

func (c *PayCallbackEvtListener) Stop() {
	c.batchConsumer.Stop()
}
