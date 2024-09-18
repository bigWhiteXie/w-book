package logic

import (
	"codexie.com/w-book-code/api/pb"
	"codexie.com/w-book-user/internal/config"
	"codexie.com/w-book-user/pkg/common/codeerr"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"testing"
)

func initCodeClient() pb.CodeClient {
	var c config.Config
	conf.MustLoad("/usr/local/go_project/w-book/app/user-service/etc/user.yaml", &c)
	return pb.NewCodeClient(zrpc.MustNewClient(c.CodeRpcConf).Conn())
}

func TestGrpcErr(t *testing.T) {
	codeClient := initCodeClient()
	_, err := codeClient.VerifyCode(context.Background(), &pb.VerifyCodeReq{Phone: "123", Biz: "login", Code: "111"})
	grpcErr := codeerr.ParseGrpcErr(err)
	fmt.Printf("code:%d, msg: %s", grpcErr.Code, grpcErr.Msg)
	fmt.Printf("kind : %T", grpcErr)
}