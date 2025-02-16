package codeerr

import (
	"context"
	"fmt"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	SystemErrCode = 000001
	UserErrCode   = 000002
)
const (
	UserEmailDuplicateCode = 100001
	UserEmailNotExistCode  = 100002
	UserPwdNotMatchCode    = 100003
	UserIdNotExistCode     = 100004
)

const (
	CodeSystemERR        = 200001
	CodeVerifyFailERR    = 200002
	CodeVerifyExcceddErr = 200003
	CodeFrequentErr      = 200004
	CodeNotExistErr      = 200006
)

const (
	SmsFrequentERR     = 300001
	SmsRecordSaveErr   = 300002
	SmsNotFoundErr     = 300003
	SmsNotAvaliableErr = 300004
)

const (
	ArtEditOtherERR = 400001
)

// 支付服务错误码
const (
	//修改数据库支付状态失败
	PayDBStatusErr = 500001
	// 支付回调发送消息失败
	PayMsgProdErr = 500002
	// 创建数据库支付记录失败
	PayDBCreateErr = 500003
	// 调用支付宝prepay接口失败
	AliPrePayErr = 500004
	// 支付宝回调签名验证失败
	AliPaySignErr = 500005
)

func LogCodeError(ctx context.Context, troubleCode string, format string, args ...interface{}) error {
	// 生成格式化错误信息
	errMsg := fmt.Sprintf(format, args...)
	_, file, line, _ := runtime.Caller(1)
	// 创建带错误码的异常
	codeErr := WithCode(SystemErrCode, errMsg) // 使用你的实际错误码常量

	// 获取调用栈并调整（跳过当前方法）
	if codeErr, ok := codeErr.(*WithCodeErr); ok {
		// 记录结构化日志
		logx.WithContext(logx.WithFields(ctx,
			logx.Field("caller", fmt.Sprintf("%s:%d", file, line)), // 排障码
			logx.Field("errCode", troubleCode),                     // 排障码
			logx.Field("cause", errMsg),                            // 错误原因
			logx.Field("stack", codeErr.StackTrace()),              // 已处理过的调用栈
		)).Error(errMsg)
	}

	return codeErr
}

func init() {
	MustRegister(SystemErrCode, 200, "系统内部错误")

	MustRegister(UserEmailDuplicateCode, 200, "该邮箱已经存在")
	MustRegister(UserEmailNotExistCode, 200, "该邮箱未注册")
	MustRegister(UserPwdNotMatchCode, 200, "密码错误")
	MustRegister(UserIdNotExistCode, 200, "用户id不存在")

	MustRegister(CodeSystemERR, 200, "验证码内部异常")
	MustRegister(CodeVerifyFailERR, 200, "验证码校验错误")
	MustRegister(CodeVerifyExcceddErr, 200, "验证码校验次数超过限制")
	MustRegister(CodeFrequentErr, 200, "验证码发送太频繁")
	MustRegister(CodeNotExistErr, 200, "验证码不存在")

	MustRegister(SmsFrequentERR, 200, "短信发送次数超过限制")
	MustRegister(SmsRecordSaveErr, 200, "短信发送失败")
	MustRegister(SmsNotFoundErr, 200, "短信不存在")
	MustRegister(SmsNotAvaliableErr, 200, "所有短信服务商均不可用，请检查网络和服务商余量")

	MustRegister(UserErrCode, 200, "用户操作越界")

}
