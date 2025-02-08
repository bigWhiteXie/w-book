package event

import (
	"context"

	"codexie.com/w-book-account/api/pb"
	"codexie.com/w-book-common/kafka/consumer"
	"codexie.com/w-book-common/middleware/filter"
	"codexie.com/w-book-reward/internal/domain"
	"codexie.com/w-book-reward/internal/logic"
	"codexie.com/w-book-reward/internal/repo"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	rewardTopic = "reward-payment-callback"

	//分成比例
	userRewardRatio = 0.95
)

type PayCallbackEvtListener struct {
	accountClient pb.AccountClient
	rewardLogic   *logic.RewardLogic
	batchConsumer *consumer.BatchConsumer[domain.PayMessage]
	bloomFilter   *filter.WindowedRedisBloomFilter
	msgRepo       repo.IMessageRepository
}

func NewPayCallbackEventListener(client sarama.Client, accountClient pb.AccountClient, rewardLogic *logic.RewardLogic, msgRepo *repo.MessageRepository, bloomFilter *filter.WindowedRedisBloomFilter) *PayCallbackEvtListener {
	l := &PayCallbackEvtListener{
		accountClient: accountClient,
		rewardLogic:   rewardLogic,
		msgRepo:       msgRepo,
		bloomFilter:   bloomFilter,
	}

	l.batchConsumer = consumer.NewBatchConsumer[domain.PayMessage](rewardTopic, client, "reward-service-group", 100, l.handleBatchPayMsg)

	return l
}

func (c *PayCallbackEvtListener) handleBatchPayMsg(eventBatch []domain.PayMessage, msgs []*sarama.ConsumerMessage) error {
	logx.WithContext(context.Background()).Infof("批量打赏支付事件,长度:%d", len(eventBatch))
	for _, msg := range eventBatch {
		exists, err := c.bloomFilter.Test(context.Background(), msg.OutTradeNo+msg.Status)
		if err != nil || exists {
			// 兜底校验
			exist, err := c.msgRepo.IsMessageProcessed(context.Background(), msg.OutTradeNo, msg.Status)
			if err != nil {
				// todo:严重告警
				logx.Errorw("幂等校验异常，等待后续补偿", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
				continue
			}

			if exist {
				logx.Infow("幂等校验不通过", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: "幂等校验不通过"})
				continue
			}
		}

		if err := c.rewardLogic.RewardCredit(context.Background(), msg.OutTradeNo, msg.Status, msg.Amt); err != nil {
			// todo:严重告警
			logx.Errorw("打赏分账异常",
				logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo},
				logx.LogField{Key: "cause", Value: err.Error()},
				logx.LogField{Key: "amt", Value: msg.Amt},
				logx.LogField{Key: "pay_status", Value: msg.Status},
			)
		}

		c.bloomFilter.Add(context.Background(), msg.OutTradeNo+msg.Status)
		if err := c.msgRepo.CreateMessage(context.Background(), msg.OutTradeNo, msg.Status); err != nil {
			// todo:告警
			logx.Errorw("打赏支付回调消息入库失败，需要补偿", logx.LogField{Key: "outTradeNo", Value: msg.OutTradeNo}, logx.LogField{Key: "cause", Value: err.Error()})
		}
	}

	return nil
}

func (c *PayCallbackEvtListener) Start() {
	logx.Infof("=========监听打赏支付回调事件=========")
	c.batchConsumer.StartListner()
}

func (c *PayCallbackEvtListener) Stop() {
	c.batchConsumer.Stop()
}
