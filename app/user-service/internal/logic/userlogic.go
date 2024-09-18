package logic

import (
	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/repo/cache"
	"codexie.com/w-book-user/internal/repo/dao"
	"codexie.com/w-book-user/internal/svc"
	"codexie.com/w-book-user/internal/types"
	"codexie.com/w-book-user/pkg/common"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"codexie.com/w-book-user/pkg/common/sql"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

type UserLogic struct {
	logx.Logger
	ctx       context.Context
	userRepo  *repo.UserRepository
	codeRpc   pb.CodeClient
	jwtSecret string
	jwtExpire int64
}

func NewUserLogic(ctx context.Context, svc *svc.ServiceContext) *UserLogic {
	userDao := dao.NewUserDao(svc.DB)
	userCache := cache.NewUserCache(svc.Cache)
	return &UserLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		userRepo:  repo.NewUserRepository(userDao, userCache),
		jwtSecret: svc.Config.Auth.AccessSecret,
		jwtExpire: svc.Config.Auth.AccessExpire,
		codeRpc:   svc.CodeRpcClient,
	}
}

func (l *UserLogic) Sign(req *types.SignReq) error {
	var pwd []byte

	pwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := &model.User{Email: sql.StringToNullString(req.Email), Password: string(pwd)}
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
	token, err := l.createTokenByUser(user)
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
	id, _ := strconv.Atoi(l.ctx.Value("id").(string))
	if user, err = l.userRepo.FindUserById(l.ctx, id); err != nil {
		return nil, err
	}
	return user, nil
}

func (l *UserLogic) SmsLogin(smsLoginReq *types.SmsLoginReq) (resp *types.LoginInfo, err error) {
	// grpc校验验证码
	codeRpcReq := &pb.VerifyCodeReq{Code: smsLoginReq.Code, Biz: "login", Phone: smsLoginReq.Phone}
	_, grpcErr := l.codeRpc.VerifyCode(l.ctx, codeRpcReq)
	if grpcErr != nil {
		return nil, codeerr.ParseGrpcErr(grpcErr)
	}
	// 根据phone查找或创建用户
	user, err := l.userRepo.FindOrCreate(l.ctx, smsLoginReq.Phone)
	if err != nil {
		return nil, err
	}
	token, err := l.createTokenByUser(user)
	// 构造token并返回
	if err != nil {
		return nil, err
	}
	return &types.LoginInfo{Token: token}, nil
}

func (l *UserLogic) SendLoginCode(req *types.SmsSendCodeReq) error {
	codeRpcReq := &pb.SendCodeReq{Biz: "login", Phone: req.Phone}
	_, grpcErr := l.codeRpc.SendCode(l.ctx, codeRpcReq)
	if grpcErr != nil {
		return codeerr.ParseGrpcErr(grpcErr)
	}
	return nil
}

func (l *UserLogic) createTokenByUser(user *model.User) (string, error) {
	claim := make(map[string]interface{})
	claim["id"] = strconv.Itoa(user.Id)
	token, err := common.GetJwtToken(l.jwtSecret, l.jwtExpire, claim)
	return token, err
}
