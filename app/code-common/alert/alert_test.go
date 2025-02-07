package alert

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerAlert(t *testing.T) {
	tests := []struct {
		name    string
		alerts  []Alert
		wantErr bool
	}{
		{
			name: "successful alert",
			alerts: []Alert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert12",
						"severity":  "critical",
						"service":   "test-service",
					},
					Annotations: map[string]string{
						"summary": "Test alert",
						"detail":  "This is a test alert",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal alert",
			alerts: []Alert{
				{
					Labels: map[string]string{
						"alertname": "MinimalTestAlert12",
						"severity":  "warning",
						"service":   "minimal-test",
					},
					Annotations: map[string]string{
						"summary": "Minimal test alert",
					},
				},
			},
			wantErr: false,
		},
	}

	// 创建 AlertService 实例
	alertService := NewAlertService(AlertConf{
		AlertManagerURL: "http://localhost:19094/api/v2/alerts",
		Timeout:         5,
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置全局 client，确保每个测试用例使用新的客户端

			// 执行测试
			err := alertService.TriggerAlert(context.Background(), tt.alerts)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test2(t *testing.T) {
	url := "http://127.0.0.1:19094/api/v2/alerts"
	method := "POST"

	payload := strings.NewReader(`[
    {
        "annotations": {
            "property1": "string",
            "property2": "string"
        },
        "labels": {
            "alertname": "MinimalTestAlert",
			"severity":  "warning",
			"service":   "minimal-test"
        },
        "generatorURL": "http://example.com"
    }
]`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
