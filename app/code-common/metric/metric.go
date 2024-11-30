package metric

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

type ConstMetricLabelsConf struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

var (
	// 消息队列指标
	ConsumeTimeHistogram *prometheus.HistogramVec
	ConsumeCountCounter  *prometheus.CounterVec
	ConsumeErrCounter    *prometheus.CounterVec

	// 短信指标
	SmsSendLatencyHistogram *prometheus.HistogramVec
	SmsSendCountCounter     *prometheus.CounterVec
	SmsSendErrCounter       *prometheus.CounterVec
	SmsConcurrentGauge      prometheus.Gauge
)

func InitMessageMetric(conf ConstMetricLabelsConf) {
	env := os.Getenv("env")
	if env == "" {
		env = "dev"
	}
	ConsumeTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "message_consumer_time_seconds",
		Help:    "Histogram of the time taken to consume Kafka messages in seconds",
		Buckets: []float64{10, 30, 50, 100, 500, 1000, 2000, 3000},
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"topic", "group"})

	ConsumeCountCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_consume_count",
		Help: "Counter for the number of Kafka messages consumed",
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"topic", "group"})

	ConsumeErrCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_consume_error_count",
		Help: "Counter for the number of Kafka messages consumed",
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"topic", "group", "cause"})
	prometheus.MustRegister(ConsumeTimeHistogram)
	prometheus.MustRegister(ConsumeCountCounter)
	prometheus.MustRegister(ConsumeErrCounter)

}

func InitSmsMetric(conf ConstMetricLabelsConf) {
	env := os.Getenv("env")
	if env == "" {
		env = "dev"
	}
	// 短信指标
	SmsSendLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sms_send_latency_seconds",
		Help:    "Histogram of the time taken to send SMS in seconds",
		Buckets: []float64{10, 30, 50, 100, 500, 1000, 2000, 3000},
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"brand"})

	SmsSendCountCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sms_send_count",
		Help: "Counter for the number of SMS sent",
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"brand"})
	SmsConcurrentGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sms_concurrent_count",
		Help: "Gauge for the number of concurrent SMS sending",
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	})
	SmsSendErrCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sms_send_error_count",
		Help: "Counter for the number of SMS sent",
		ConstLabels: prometheus.Labels{
			"version": conf.Version,
			"env":     env,
		},
	}, []string{"brand", "cause"})
	prometheus.MustRegister(SmsSendLatencyHistogram)
	prometheus.MustRegister(SmsSendCountCounter)
	prometheus.MustRegister(SmsSendErrCounter)
	prometheus.MustRegister(SmsConcurrentGauge)
}
