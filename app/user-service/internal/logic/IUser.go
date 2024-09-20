package logic

import (
	"context"

	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/types"
)

// mockgen -source=internal/logic/IUser.go -destination mocks/logic/user_mock.go
type IUserLogic interface {
	Sign(ctx context.Context, req *types.SignReq) error
	Login(ctx context.Context, req *types.LoginReq) (resp *types.LoginInfo, err error)
	Edit(ctx context.Context, req *types.UserInfoReq) error
	Profile(ctx context.Context) (user *model.User, err error)
	SmsLogin(ctx context.Context, smsLoginReq *types.SmsLoginReq) (resp *types.LoginInfo, err error)
	SendLoginCode(ctx context.Context, req *types.SmsSendCodeReq) error
}
