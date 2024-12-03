package producer

import (
	"context"
)

// ProducerOption 选项配置函数类型
type ProducerOption func(*ProducerOptions)

// ProducerOptions 包含可选参数，如 Key 和 Headers
type ProducerOptions struct {
	Key     string
	Headers map[string]string
}

// Producer 消息队列生产者接口
type Producer interface {
	SendSync(ctx context.Context, topic string, msg string, opts ...ProducerOption) error
	SendAsync(ctx context.Context, topic string, msg string, onError func(error), opts ...ProducerOption)
}

// ApplyOptions 应用可选参数
func ApplyOptions(opts ...ProducerOption) *ProducerOptions {
	options := &ProducerOptions{
		Headers: make(map[string]string),
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithKey 设置消息的 Key
func WithKey(key string) ProducerOption {
	return func(opts *ProducerOptions) {
		opts.Key = key
	}
}

// WithHeader 添加 Header
func WithHeader(key, value string) ProducerOption {
	return func(opts *ProducerOptions) {
		opts.Headers[key] = value
	}
}
