package handler

import (
	"net/http"

	"codexie.com/w-book-payment/internal/logic"
	"codexie.com/w-book-payment/internal/svc"
	"codexie.com/w-book-payment/internal/types"
	"codexie.com/w-book-payment/pkg/constant"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AliPayHandler struct {
	ctx           *svc.ServiceContext
	interactLogic *logic.AliPayLogic
}

func NewAliPayHandler(ctx *svc.ServiceContext, alipayLogic *logic.AliPayLogic) *AliPayHandler {
	return &AliPayHandler{
		ctx:           ctx,
		interactLogic: alipayLogic,
	}
}

func (h *AliPayHandler) Callback(w http.ResponseWriter, r *http.Request) {
	var req types.AliPaymentMsg
	values := r.URL.Query()
	// 将 URL 参数映射到结构体
	req.OutTradeNo = values.Get("out_trade_no")
	req.TradeStatus = values.Get("trade_status")
	// ... 其他字段映射
	if err := httpx.ParseForm(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	payLogic, _ := h.ctx.GetPayLogic(constant.ZFB)
	if err := payLogic.VerifySign(r); err != nil {
		httpx.Error(w, err)
		return
	}
	payLogic.PayCallback(r.Context(), &req)

	httpx.Ok(w)
}
