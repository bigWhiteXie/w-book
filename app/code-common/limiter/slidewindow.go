package limiter

import (
	"context"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	wsize  time.Duration //窗口大小
	limit  int64         //窗口容量
	window []time.Time   //窗口
	mu     sync.RWMutex
}

func NewSlideWindowLimiter(wsize time.Duration, limit int64) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		wsize:  wsize,
		limit:  limit,
		window: make([]time.Time, 0),
	}
}

func (l *SlideWindowLimiter) Allow(ctx context.Context) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.window) < int(l.limit) {
		l.window = append(l.window, now)
		return true
	}

	begin := now.Truncate(l.wsize)
	index := 0
	for index < len(l.window) {
		if l.window[index].After(begin) {
			break
		}
		index++
	}
	if index == 0 {
		return false
	}
	l.window = append(l.window[index:], now)

	return true
}
