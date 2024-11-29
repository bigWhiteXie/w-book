package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"codexie.com/w-book-code/pkg/sms"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	endpoint = "sms.tencentcloudapi.com"
	sign     = "HmacSHA1" //考虑使用环境变量
)

type TCSmsRequst struct {
	TemplateId       string   `json`
	TemplateParamSet []string `json`
}

type TCSmsProvider struct {
	appId        string
	signName     string
	status       int
	weight       int
	failCount    int
	smsClient    *tcsms.Client
	lastFailTime time.Time
}

func NewTCSmsClient(conf sms.Tencent) SmsProvider {
	credential := common.NewCredential(
		conf.SecretId,
		conf.SecretKey)

	/* 实例化一个客户端配置对象，可以指定超时时间等配置 */
	cpf := profile.NewClientProfile()

	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10 // 请求超时时间，单位为秒(默认60秒)
	cpf.HttpProfile.Endpoint = endpoint
	cpf.SignMethod = sign

	smsClient, err := tcsms.NewClient(credential, conf.Region, cpf)
	if err != nil {
		logx.Errorf("[TcSmsProvider] 初始化client失败, 原因:%s", err)
		panic(err)
	}

	return &TCSmsProvider{
		appId:     conf.AppId,
		signName:  conf.SignName,
		smsClient: smsClient,
		weight:    conf.Weight,
		status:    Avaliable,
	}
}

func (p *TCSmsProvider) SendSms(ctx context.Context, phone string, args map[string]string) error {
	request, err := p.mapToTCSmsRequst(args)
	if err != nil {
		return err
	}

	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := p.smsClient.SendSms(request)
	// 处理异常
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logx.Errorf("[TCSmsProvider] 发送短信失败: %s", err)
		return err
	}

	b, err := json.Marshal(response.Response)
	if err != nil {
		logx.Errorf("[TCSmsProvider] 反序列化响应内容失败:%s", err)
		return nil
	}
	logx.Infof("[TCSmsProvider] 发送短信响应内容:%s", b)
	return nil
}

func (client *TCSmsProvider) GetName() string {
	return "tencent-sms"
}

func (client *TCSmsProvider) GetWeight() int {
	return client.weight
}

func (client *TCSmsProvider) GetStatus() int {
	return client.status
}

func (client *TCSmsProvider) GetFailCount() int {
	return client.failCount
}

func (client *TCSmsProvider) GetFailTime() time.Time {
	return client.lastFailTime
}

func (client *TCSmsProvider) SetFailTime(t time.Time) {
	client.lastFailTime = t
}

func (client *TCSmsProvider) SetFailCount(count int) {
	client.failCount = count
}

func (client *TCSmsProvider) SetWeight(weight int) {
	client.weight = weight
}

func (client *TCSmsProvider) SetStatus(status int) {
	client.status = status
}

func (client *TCSmsProvider) mapToTCSmsRequst(args map[string]string) (*tcsms.SendSmsRequest, error) {
	// 创建一个 TCSmsRequst 实例
	req := &tcsms.SendSmsRequest{}

	// 从 map 中提取值并进行类型断言和转换
	req.SmsSdkAppId = common.StringPtr(client.appId)
	req.SignName = common.StringPtr(client.signName)

	if templateId, ok := args["TemplateId"]; ok {
		req.TemplateId = common.StringPtr(templateId)
	} else {
		return nil, fmt.Errorf("invalid type for TemplateId")
	}

	if templateParam, ok := args["TemplateParamSet"]; ok {
		// 将 []interface{} 转换为 []string
		templateParamSet := strings.Split(templateParam, ",")
		for _, param := range templateParamSet {
			req.TemplateParamSet = append(req.TemplateParamSet, common.StringPtr(param))
		}
	} else {
		return nil, fmt.Errorf("invalid type for TemplateParamSet")
	}

	return req, nil
}
