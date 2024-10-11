package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"codexie.com/w-book-user/internal/logic"
	"codexie.com/w-book-user/internal/model"
	"codexie.com/w-book-user/internal/types"
	mock_logic "codexie.com/w-book-user/mocks/logic"
	"github.com/golang/mock/gomock"
)

func TestUserHandler_EditHandler(t *testing.T) {
	type fields struct {
		userLogic logic.IUserLogic
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserHandler{
				userLogic: tt.fields.userLogic,
			}
			u.EditHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestUserHandler_LoginHandler(t *testing.T) {

	ctrl := gomock.NewController(t)
	type fields struct {
		userLogic logic.IUserLogic
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		mock   func(ctrl *gomock.Controller) logic.IUserLogic
		fields fields
		args   args
	}{
		{
			name: "正常登录",
			mock: func(ctrl *gomock.Controller) logic.IUserLogic {
				userLogic := mock_logic.NewMockIUserLogic(ctrl)
				userLogic.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&model.User{}, nil)
				return userLogic
			},
		},
		{
			name: "异常登录登录",
			mock: func(ctrl *gomock.Controller) logic.IUserLogic {
				userLogic := mock_logic.NewMockIUserLogic(ctrl)
				userLogic.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.New("异常登录"))
				return userLogic
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserHandler{
				userLogic: tt.mock(ctrl),
			}
			tt.args.w = httptest.NewRecorder()
			req := types.LoginReq{
				Email:    "example@example.com",
				Password: "mypassword",
			}

			// 将结构体转换为 JSON
			jsonData, err := json.Marshal(req)
			if err != nil {
				log.Fatalf("Error marshalling to JSON: %v", err)
			}

			// 将 JSON 数据转换为 io.Reader 类型
			reader := bytes.NewBuffer(jsonData)
			tt.args.r, _ = http.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/user/login", reader)
			tt.args.r.Header.Set("Content-Type", "application/json")
			u.LoginHandler(tt.args.w, tt.args.r)
		})
	}
}
