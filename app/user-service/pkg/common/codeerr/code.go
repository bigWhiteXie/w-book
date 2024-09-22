package codeerr

const (
	SystemErrCpde = 000001
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

func init() {
	MustRegister(SystemErrCpde, 200, "系统内部错误")

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
}
