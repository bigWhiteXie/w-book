package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

var (
	endpoint = "sms.tencentcloudapi.com"
	sign     = "HmacSHA1"
)

var smsClient *tcsms.Client

type TCSmsRequst struct {
	TemplateId       string   `json`
	TemplateParamSet []string `json`
}

type TCSmsClient struct {
	appId    string
	signName string
}

func NewTCSmsClient(conf Tencent) *TCSmsClient {
	credential := common.NewCredential(
		conf.SecretId,
		conf.SecretKey)

	/* 实例化一个客户端配置对象，可以指定超时时间等配置 */
	cpf := profile.NewClientProfile()

	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10 // 请求超时时间，单位为秒(默认60秒)
	cpf.HttpProfile.Endpoint = endpoint
	cpf.SignMethod = sign

	smsClient, _ = tcsms.NewClient(credential, conf.Region, cpf)

	return &TCSmsClient{
		appId:    conf.AppId,
		signName: conf.SignName,
	}
}

func (client *TCSmsClient) SendSms(ctx context.Context, phone string, args map[string]string) error {
	request, err := client.mapToTCSmsRequst(args)
	if err != nil {
		return err
	}

	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := smsClient.SendSms(request)
	// 处理异常
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}

	b, _ := json.Marshal(response.Response)
	// 打印返回的json字符串
	fmt.Printf("短信发送：%s", b)
	return nil
}

func (client *TCSmsClient) mapToTCSmsRequst(args map[string]string) (*tcsms.SendSmsRequest, error) {
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
