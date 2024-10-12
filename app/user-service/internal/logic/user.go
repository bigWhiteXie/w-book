package logic

// mockgen -source=./app/user-service/internal/logic/user.go -package=svcmocks destination=./app/user-service/internal/logic/mocks/user.mock.go
import (
	"context"
	"strconv"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-common/common"
	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-common/common/sql"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserLogic struct {
	logx.Logger
	userRepo  repo.IUserRepository
	codeRpc   pb.CodeClient
	jwtSecret string
	jwtExpire int64
}

func NewUserLogic(c config.Config, userRepo repo.IUserRepository, codeRpc pb.CodeClient) IUserLogic {
	return &UserLogic{
		userRepo:  userRepo,
		jwtSecret: c.Auth.AccessSecret,
		jwtExpire: c.Auth.AccessExpire,
		codeRpc:   codeRpc,
	}
}

func (l *UserLogic) Sign(ctx context.Context, req *types.SignReq) error {
	var pwd []byte

	pwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := &model.User{Email: sql.StringToNullString(req.Email), Password: string(pwd)}
	if err = l.userRepo.Create(ctx, user); err != nil {
		return err
	}
	return nil
}

func (l *UserLogic) Login(ctx context.Context, req *types.LoginReq) (resp *model.User, err error) {
	var user *model.User
	if user, err = l.userRepo.FindUserByEmail(ctx, req.Email); err != nil {
		return nil, codeerr.WithCode(codeerr.UserEmailNotExistCode, "[Login] 邮箱 %s 不存在", req.Email)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, codeerr.WithCode(codeerr.UserPwdNotMatchCode, "[Login] 用户id %s 密码错误", user.Id)
	}

	return user, nil
}

func (l *UserLogic) Edit(ctx context.Context, req *types.UserInfoReq) error {
	//email := l.ctx.Value("email").(string)
	//user := &model.User{Email: email, Password: req.Password}
	return nil
}

func (l *UserLogic) Profile(ctx context.Context) (user *model.User, err error) {
	id, _ := strconv.Atoi(ctx.Value("id").(string))
	if user, err = l.userRepo.FindUserById(ctx, id); err != nil {
		return nil, err
	}
	return user, nil
}

func (l *UserLogic) SmsLogin(ctx context.Context, smsLoginReq *types.SmsLoginReq) (resp *model.User, err error) {
	// grpc校验验证码
	codeRpcReq := &pb.VerifyCodeReq{Code: smsLoginReq.Code, Biz: "login", Phone: smsLoginReq.Phone}
	_, grpcErr := l.codeRpc.VerifyCode(ctx, codeRpcReq)
	if grpcErr != nil {
		return nil, codeerr.ParseGrpcErr(grpcErr)
	}
	// 根据phone查找或创建用户
	user, err := l.userRepo.FindOrCreate(ctx, smsLoginReq.Phone)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (l *UserLogic) SendLoginCode(ctx context.Context, req *types.SmsSendCodeReq) error {
	codeRpcReq := &pb.SendCodeReq{Biz: "login", Phone: req.Phone}
	_, grpcErr := l.codeRpc.SendCode(ctx, codeRpcReq)
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
