package handler

import (
	"net/http"

	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/types"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"codexie.com/w-book-user/pkg/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserHandler struct {
	userLogic logic.IUserLogic
}

func NewUserHandler(userLogic logic.IUserLogic) *UserHandler {
	return &UserHandler{
		userLogic: userLogic,
	}
}

func (u *UserHandler) EditHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.UserInfoReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := u.userLogic.Edit(r.Context(), &req)
	if err != nil {
		resp = response.Fail(500, "系统错误，注册失败")
	} else {
		resp = response.Ok(nil)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (u *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.LoginReq
		resp *response.Response
	)
	if err := httpx.ParseJsonBody(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	loginInfo, err := u.userLogic.Login(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(loginInfo)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (u *UserHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	var resp *response.Response
	user, err := u.userLogic.Profile(r.Context())
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(user)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (u *UserHandler) SignHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.SignReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := u.userLogic.Sign(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (u *UserHandler) SmsLoginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.SmsLoginReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	//todo:校验参数

	loginInfo, err := u.userLogic.SmsLogin(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(loginInfo)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (u *UserHandler) SendLoginCodeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  types.SmsSendCodeReq
		resp *response.Response
	)
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	//todo:校验参数
	err := u.userLogic.SendLoginCode(r.Context(), &req)
	if err != nil {
		resp = codeerr.HandleErr(r.Context(), err)
	} else {
		resp = response.Ok(nil)
	}
	httpx.OkJsonCtx(r.Context(), w, resp)
}
