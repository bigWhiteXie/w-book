package logic

import (
	"context"
	"reflect"
	"testing"

	"codexie.com/w-book-code/api/pb"
	mock_pb "codexie.com/w-book-code/mocks/api/pb"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/types"
	mock_repo "codexie.com/w-book-user/mocks/repo"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"codexie.com/w-book-user/pkg/common/sql"
	"github.com/golang/mock/gomock"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

func TestUserLogic_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	conf := config.Config{
		Auth: struct {
			AccessSecret string
			AccessExpire int64
		}{
			AccessSecret: "123456",
			AccessExpire: 600,
		},
	}
	type fields struct {
		Logger    logx.Logger
		userRepo  repo.IUserRepository
		codeRpc   pb.CodeClient
		jwtSecret string
		jwtExpire int64
	}
	type args struct {
		ctx context.Context
		req *types.LoginReq
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mock     func(ctrl *gomock.Controller) *UserLogic
		wantResp *types.LoginInfo
		wantErr  bool
	}{
		{
			name: "正常登录",
			args: args{
				ctx: context.Background(),
				req: &types.LoginReq{
					Email:    "2607219580@qq.com",
					Password: "123456",
				},
			},
			mock: func(ctrl *gomock.Controller) *UserLogic {
				userRepo := mock_repo.NewMockIUserRepository(ctrl)
				pwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
				userRepo.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Return(&model.User{Id: 1, Email: sql.StringToNullString("2607219580@qq.com"), Password: string(pwd)}, nil)
				return NewUserLogic(conf, userRepo, mock_pb.NewMockCodeClient(ctrl)).(*UserLogic)
			},
			wantErr: false,
		},
		{
			name: "邮箱不存在",
			args: args{
				ctx: context.Background(),
				req: &types.LoginReq{
					Email:    "2607219580@qq.com",
					Password: "123456",
				},
			},
			mock: func(ctrl *gomock.Controller) *UserLogic {
				userRepo := mock_repo.NewMockIUserRepository(ctrl)
				userRepo.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Return(nil, codeerr.WithCode(codeerr.UserEmailNotExistCode, "can't find any user by email %s", "2607219580@qq.com"))
				return NewUserLogic(conf, userRepo, mock_pb.NewMockCodeClient(ctrl)).(*UserLogic)
			},
			wantErr:  true,
			wantResp: nil,
		},
		{
			name: "密码不正确",
			args: args{
				ctx: context.Background(),
				req: &types.LoginReq{
					Email:    "2607219580@qq.com",
					Password: "123456",
				},
			},
			mock: func(ctrl *gomock.Controller) *UserLogic {
				userRepo := mock_repo.NewMockIUserRepository(ctrl)
				pwd, _ := bcrypt.GenerateFromPassword([]byte("1234567"), bcrypt.DefaultCost)
				userRepo.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Return(&model.User{Id: 1, Email: sql.StringToNullString("2607219580@qq.com"), Password: string(pwd)}, nil)
				return NewUserLogic(conf, userRepo, mock_pb.NewMockCodeClient(ctrl)).(*UserLogic)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.mock(ctrl)

			loginToken, _ := l.createTokenByUser(&model.User{Id: 1, Email: sql.StringToNullString("2607219580@qq.com"), Password: "$2a$10$v.U6vL7HZUyTSufF2Qbbke2P8nOo38vyxr7tLENEsViBT2ZuoDV2y"})

			tt.wantResp = &types.LoginInfo{Token: loginToken}
			gotResp, err := l.Login(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserLogic.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				t.Logf("error: %v", err)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("UserLogic.Login() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
