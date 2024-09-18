package handler

import (
	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/svc"
	"codexie.com/w-book-user/internal/types"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"codexie.com/w-book-user/pkg/common/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func EditHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  types.UserInfoReq
			resp *response.Response
		)
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserLogic(r.Context(), svcCtx)
		err := l.Edit(&req)
		if err != nil {
			resp = response.Fail(500, "系统错误，注册失败")
		} else {
			resp = response.Ok(nil)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  types.LoginReq
			resp *response.Response
		)
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserLogic(r.Context(), svcCtx)
		loginInfo, err := l.Login(&req)
		if err != nil {
			resp = codeerr.HandleErr(r.Context(), err)
		} else {
			resp = response.Ok(loginInfo)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)

	}
}

func ProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp *response.Response
		l := logic.NewUserLogic(r.Context(), svcCtx)
		user, err := l.Profile()
		if err != nil {
			resp = codeerr.HandleErr(r.Context(), err)
		} else {
			resp = response.Ok(user)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func SignHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  types.SignReq
			resp *response.Response
		)
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserLogic(r.Context(), svcCtx)
		err := l.Sign(&req)
		if err != nil {
			resp = codeerr.HandleErr(r.Context(), err)
		} else {
			resp = response.Ok(nil)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func SmsLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  types.SmsLoginReq
			resp *response.Response
		)
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		//todo:校验参数

		l := logic.NewUserLogic(r.Context(), svcCtx)
		loginInfo, err := l.SmsLogin(&req)
		if err != nil {
			resp = codeerr.HandleErr(r.Context(), err)
		} else {
			resp = response.Ok(loginInfo)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func SendLoginCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  types.SmsSendCodeReq
			resp *response.Response
		)
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		//todo:校验参数

		l := logic.NewUserLogic(r.Context(), svcCtx)
		err := l.SendLoginCode(&req)
		if err != nil {
			resp = codeerr.HandleErr(r.Context(), err)
		} else {
			resp = response.Ok(nil)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
