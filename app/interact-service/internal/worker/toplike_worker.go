package worker

import (
	"context"
	"time"

	"codexie.com/w-book-interact/internal/repo"
	"github.com/zeromicro/go-zero/core/logx"
)

type TopLikeWorker struct {
	frequentce   time.Duration
	interactRepo repo.IInteractRepo
	times        int
}

func NewTopLikeWorker(interactRepo repo.IInteractRepo) Worker {
	return &TopLikeWorker{
		frequentce:   60 * time.Second,
		interactRepo: interactRepo,
	}
}

func (w *TopLikeWorker) Init() {
	for _, biz := range resourceTypes {
		if err := w.interactRepo.RefreshTopLikeRedis(context.Background(), biz, 500); err != nil {
			logx.Errorf("更新资源[%s]的缓存失败,原因:%s", biz, err)
		}
		w.interactRepo.RefreshTopLikeLocal(context.Background(), biz, 100)
	}
}

func (w *TopLikeWorker) DoWork() {
	w.times++
	for _, biz := range resourceTypes {
		if w.times%5 == 0 {
			err := w.interactRepo.RefreshTopLikeRedis(context.Background(), biz, 500)
			if err != nil {
				logx.Errorf("更新资源[%s]的缓存失败,原因:%s", biz, err)
			}
		}
		w.interactRepo.RefreshTopLikeLocal(context.Background(), biz, 100)
	}
	w.times = w.times % 5
}

func (w *TopLikeWorker) GetFrenquency() time.Duration {
	return w.frequentce
}
