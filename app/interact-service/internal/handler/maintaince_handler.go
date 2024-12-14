package handler

import (
	"fmt"
	"net/http"

	"codexie.com/w-book-common/repo"
	"github.com/zeromicro/go-zero/rest/httpx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbChangeReq struct {
	DSN    string `json:"dsn"`
	Driver string `json:"driver"`
}

type MaintainceHandler struct {
	baseRepo *repo.BaseRepo
}

func NewMaintainceHandler(baseRepo *repo.BaseRepo) *MaintainceHandler {
	return &MaintainceHandler{
		baseRepo: baseRepo,
	}
}

func (h *MaintainceHandler) ChangeDB(w http.ResponseWriter, r *http.Request) {
	var (
		req DbChangeReq
		db  *gorm.DB
		err error
	)

	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	dsn := req.DSN
	switch req.Driver {
	case "mysql", "tidb":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		panic(fmt.Sprintf("unsupported database driver: %s", req.Driver))
	}
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	h.baseRepo.SetDB(db)
	httpx.OkJsonCtx(r.Context(), w, "ok")
}
