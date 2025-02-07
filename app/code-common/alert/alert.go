package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlertConf struct {
	AlertManagerURL string `json:"alertManagerURL"` // AlertManager 地址
	Timeout         int    `json:"timeout"`         // 请求超时时间（秒）
}

// AlertService 告警配置
type AlertService struct {
	AlertConf // 请求超时时间（秒）
}

// Alert 告警结构体
type Alert struct {
	Labels      map[string]string `json:"labels"`      // 告警标签
	Annotations map[string]string `json:"annotations"` // 告警注释
	StartsAt    time.Time         `json:"startsAt"`    // 告警开始时间
}

// AlertManagerRequest AlertManager 请求结构体
type AlertManagerRequest struct {
	Alerts []Alert `json:"alerts"`
}

var (
	client     *http.Client
	clientOnce sync.Once
)

// InitAlertConf 创建告警配置
func NewAlertService(conf AlertConf) *AlertService {
	return &AlertService{
		AlertConf: conf,
	}
}

// getClient 获取单例的 HTTP 客户端
func getClient(timeout int) *http.Client {
	clientOnce.Do(func() {
		client = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	})
	return client
}

// TriggerAlert 触发告警
func (a *AlertService) TriggerAlert(ctx context.Context, alerts []Alert) error {
	// 构建请求体
	reqBody := AlertManagerRequest{
		Alerts: alerts,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logx.Errorf("序列化告警请求失败: %v", err)
		return err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", a.AlertManagerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logx.Errorf("创建 HTTP 请求失败: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 获取单例的 HTTP 客户端
	client := getClient(a.Timeout)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logx.Errorf("发送告警请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		logx.Errorf("告警请求失败，状态码: %d", resp.StatusCode)
		return err
	}

	logx.Info("告警触发成功")
	return nil
}
