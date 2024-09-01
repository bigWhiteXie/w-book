package common

import "errors"

var (
	UserEmailDuplicateErr = errors.New("该邮箱已经存在")
	UserEmailNotExistErr  = errors.New("邮箱不存在")
	UserPwdNotMatchErr    = errors.New("密码错误")
)
