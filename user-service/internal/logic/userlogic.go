package logic

import (
	"codexie.com/w-book-user/internal/common"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/internal/svc"
	"codexie.com/w-book-user/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserLogic struct {
	logx.Logger
	ctx       context.Context
	userRepo  *repo.UserRepository
	jwtSecret string
	jwtExpire int64
}

func NewUserLogic(ctx context.Context, svc *svc.ServiceContext) *UserLogic {
	userDao := dao.NewUserDao(svc.DB)
	return &UserLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		userRepo:  repo.NewUserRepository(userDao),
		jwtSecret: svc.Config.Auth.AccessSecret,
		jwtExpire: svc.Config.Auth.AccessExpire,
	}
}

func (l *UserLogic) Sign(req *types.SignReq) error {
	var pwd []byte
	pwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := &model.User{Email: req.Email, Password: string(pwd)}
	if err = l.userRepo.Create(l.ctx, user); err != nil {
		return err
	}
	return nil
}

func (l *UserLogic) Login(req *types.LoginReq) (resp *types.LoginInfo, err error) {
	var user *model.User
	if user, err = l.userRepo.FindUserByEmail(l.ctx, req.Email); err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, err
	}
	claim := make(map[string]interface{})
	claim["email"] = req.Email
	token, err := common.GetJwtToken(l.jwtSecret, l.jwtExpire, claim)
	if err != nil {
		return nil, err
	}
	return &types.LoginInfo{Token: token}, nil
}

func (l *UserLogic) Edit(req *types.UserInfoReq) error {
	//email := l.ctx.Value("email").(string)
	//user := &model.User{Email: email, Password: req.Password}
	return nil
}

func (l *UserLogic) Profile() (user *model.User, err error) {
	email := l.ctx.Value("email").(string)
	if user, err = l.userRepo.FindUserByEmail(l.ctx, email); err != nil {
		return nil, err
	}
	return user, nil
}
