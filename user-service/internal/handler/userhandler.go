package handler

import (
	"codexie.com/w-book-user/internal/common"
	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/svc"
	"codexie.com/w-book-user/internal/types"
	"codexie.com/w-book-user/internal/types/response"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
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
			logx.Errorf("fail to login, cause: %s", err.Error())
			if errors.Is(err, common.UserPwdNotMatchErr) {
				resp = response.Fail(500, "密码错误")
			}
			resp = response.Fail(500, "系统错误，注册失败")
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
			logx.Errorf("fail to get user profile, cause: %s", err.Error())
			if errors.Is(err, common.UserEmailNotExistErr) {
				resp = response.Fail(500, "该邮箱不存在")
			}
			resp = response.Fail(500, "系统错误，注册失败")
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
			if errors.Is(err, common.UserEmailDuplicateErr) {
				logx.Errorf("fail to sign user, cause: %s", err.Error())
				resp = response.Fail(500, "该邮箱已被注册")
			}
			resp = response.Fail(500, "系统错误，注册失败")

		} else {
			resp = response.Ok(nil)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
