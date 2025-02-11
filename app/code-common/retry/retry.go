package retry

import (
	"time"

	"github.com/zeromicro/go-zero/core/threading"
)

func AsyncRetry(interval time.Duration, MaxRetries int, fn func() error, alertFunc func(err error)) {
	threading.GoSafe(func() {
		var lastErr error
		for i := 0; i < MaxRetries; i++ {
			if i > 0 {
				select {
				case <-time.After(interval): // 上下文取消
				}
			}

			if err := fn(); err != nil {
				lastErr = err
				continue
			}
			return // 成功则退出
		}

		// todo:全部重试失败进行告警
		alertFunc(lastErr)
	})
}
