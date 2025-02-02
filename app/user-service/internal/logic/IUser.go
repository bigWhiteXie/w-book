package logic

import (
	"context"

	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/types"
)

//go:generate mockgen -source=./app/user-service/internal/logic/user.go -package=svcmocks -destination=./app/user-service/internal/logic/mocks/user.mock.go

// mockgen -source=internal/logic/IUser.go -destination mocks/logic/user_mock.go
type IUserLogic interface {
	Sign(ctx context.Context, req *types.SignReq) error
	Login(ctx context.Context, req *types.LoginReq) (resp *model.User, err error)
	Edit(ctx context.Context, req *types.UserInfoReq) error
	Profile(ctx context.Context) (user *model.User, err error)
	SmsLogin(ctx context.Context, smsLoginReq *types.SmsLoginReq) (resp *model.User, err error)
	SendLoginCode(ctx context.Context, req *types.SmsSendCodeReq) error
	GenerateWechatLoginQR() (string, string, error)
	HandleWechatCallback(ctx context.Context, code string) (*model.User, error)
}
