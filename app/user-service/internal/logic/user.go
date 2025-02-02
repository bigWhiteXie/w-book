package logic

// mockgen -source=./app/user-service/internal/logic/user.go -package=svcmocks destination=./app/user-service/internal/logic/mocks/user.mock.go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-common/codeerr"
	"codexie.com/w-book-common/ijwt"
	"codexie.com/w-book-common/sql"

	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/repo"
	"codexie.com/w-book-user/internal/types"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type WechatTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

type WechatUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

type UserLogic struct {
	logx.Logger
	userRepo  repo.IUserRepository
	codeRpc   pb.CodeClient
	jwtSecret string
	jwtExpire int64
}

func NewUserLogic(c config.Config, userRepo repo.IUserRepository, codeRpc pb.CodeClient) *UserLogic {
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
		return nil, errors.WithMessage(err, "[UserLogic_Login] 根据邮箱查找用户失败")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.Wrap(codeerr.WithCode(codeerr.UserPwdNotMatchCode, "[UserLogic_Login] 邮箱=%s,密码错误:%s", req.Email, err), "")
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
		return errors.Wrap(codeerr.ParseGrpcErr(grpcErr), "")
	}
	return nil
}

func (l *UserLogic) createTokenByUser(user *model.User) (string, error) {
	claim := make(map[string]interface{})
	claim["id"] = strconv.Itoa(user.Id)
	token, err := ijwt.GetJwtToken(l.jwtSecret, l.jwtExpire, claim)
	return token, err
}

func (l *UserLogic) GenerateWechatLoginQR() (string, string, error) {
	// 生成随机state参数，用于防止CSRF攻击
	state := uuid.New().String()

	// 构造微信登录URL
	qrCodeURL := fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?"+
		"appid=%s"+
		"&redirect_uri=%s"+
		"&response_type=code"+
		"&scope=snsapi_login"+
		"&state=%s",
		url.QueryEscape("appid"),
		url.QueryEscape("redirect_uri"),
		url.QueryEscape(state),
	)

	return qrCodeURL, state, nil
}

func (l *UserLogic) HandleWechatCallback(ctx context.Context, code string) (*model.User, error) {
	// 通过code获取access_token
	tokenResp, err := l.getWechatAccessToken(code)
	if err != nil {
		return nil, errors.Wrap(err, "获取微信access_token失败")
	}

	// 获取用户信息
	userInfo, err := l.getWechatUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		return nil, errors.Wrap(err, "获取微信用户信息失败")
	}

	// 查找或创建用户
	user, err := l.userRepo.FindOrCreateByWechat(ctx, userInfo.OpenID, userInfo.Nickname, userInfo.HeadImgURL)
	if err != nil {
		return nil, errors.Wrap(err, "创建用户失败")
	}

	return user, nil
}

// 获取微信access_token
func (l *UserLogic) getWechatAccessToken(code string) (*WechatTokenResp, error) {
	// 构造获取access_token的URL
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		"appid",
		"appsecret",
		code,
	)

	// 发送请求
	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("请求access_token失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var tokenResp WechatTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("解析access_token响应失败: %v", err)
	}

	// 检查是否返回了错误
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("获取access_token失败")
	}

	return &tokenResp, nil
}

func (l *UserLogic) getWechatUserInfo(accessToken, openID string) (*WechatUserInfo, error) {
	// 构造获取用户信息的URL
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken,
		openID,
	)

	// 发送请求
	resp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("请求用户信息失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var userInfo WechatUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息响应失败: %v", err)
	}

	return &userInfo, nil
}
