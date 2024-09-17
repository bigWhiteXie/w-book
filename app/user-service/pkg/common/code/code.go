package code

import "codexie.com/w-book-user/pkg/common/codeerr"

const (
	UserEmailDuplicateCode = 100001
	UserEmailNotExistCode  = 100002
	UserPwdNotMatchCode    = 100003
	UserIdNotExistCode     = 100004
)

func init() {
	codeerr.MustRegister(UserEmailDuplicateCode, 200, "该邮箱已经存在")
	codeerr.MustRegister(UserEmailNotExistCode, 200, "该邮箱未注册")
	codeerr.MustRegister(UserPwdNotMatchCode, 200, "密码错误")
	codeerr.MustRegister(UserIdNotExistCode, 200, "用户id不存在")
}
