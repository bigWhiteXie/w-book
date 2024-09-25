// Code generated by MockGen. DO NOT EDIT.
// Source: internal/logic/IUser.go

// Package mock_logic is a generated GoMock package.
package mock_logic

import (
	context "context"
	reflect "reflect"

	model "codexie.com/w-book-user/internal/model"
	types "codexie.com/w-book-user/internal/types"
	gomock "github.com/golang/mock/gomock"
)

// MockIUserLogic is a mock of IUserLogic interface.
type MockIUserLogic struct {
	ctrl     *gomock.Controller
	recorder *MockIUserLogicMockRecorder
}

// MockIUserLogicMockRecorder is the mock recorder for MockIUserLogic.
type MockIUserLogicMockRecorder struct {
	mock *MockIUserLogic
}

// NewMockIUserLogic creates a new mock instance.
func NewMockIUserLogic(ctrl *gomock.Controller) *MockIUserLogic {
	mock := &MockIUserLogic{ctrl: ctrl}
	mock.recorder = &MockIUserLogicMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIUserLogic) EXPECT() *MockIUserLogicMockRecorder {
	return m.recorder
}

// Edit mocks base method.
func (m *MockIUserLogic) Edit(ctx context.Context, req *types.UserInfoReq) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Edit", ctx, req)
	ret0, _ := ret[0].(error)
	return ret0
}

// Edit indicates an expected call of Edit.
func (mr *MockIUserLogicMockRecorder) Edit(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Edit", reflect.TypeOf((*MockIUserLogic)(nil).Edit), ctx, req)
}

// Login mocks base method.
func (m *MockIUserLogic) Login(ctx context.Context, req *types.LoginReq) (*types.LoginInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, req)
	ret0, _ := ret[0].(*types.LoginInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockIUserLogicMockRecorder) Login(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockIUserLogic)(nil).Login), ctx, req)
}

// Profile mocks base method.
func (m *MockIUserLogic) Profile(ctx context.Context) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Profile", ctx)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Profile indicates an expected call of Profile.
func (mr *MockIUserLogicMockRecorder) Profile(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Profile", reflect.TypeOf((*MockIUserLogic)(nil).Profile), ctx)
}

// SendLoginCode mocks base method.
func (m *MockIUserLogic) SendLoginCode(ctx context.Context, req *types.SmsSendCodeReq) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendLoginCode", ctx, req)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendLoginCode indicates an expected call of SendLoginCode.
func (mr *MockIUserLogicMockRecorder) SendLoginCode(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendLoginCode", reflect.TypeOf((*MockIUserLogic)(nil).SendLoginCode), ctx, req)
}

// Sign mocks base method.
func (m *MockIUserLogic) Sign(ctx context.Context, req *types.SignReq) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sign", ctx, req)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sign indicates an expected call of Sign.
func (mr *MockIUserLogicMockRecorder) Sign(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sign", reflect.TypeOf((*MockIUserLogic)(nil).Sign), ctx, req)
}

// SmsLogin mocks base method.
func (m *MockIUserLogic) SmsLogin(ctx context.Context, smsLoginReq *types.SmsLoginReq) (*types.LoginInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SmsLogin", ctx, smsLoginReq)
	ret0, _ := ret[0].(*types.LoginInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SmsLogin indicates an expected call of SmsLogin.
func (mr *MockIUserLogicMockRecorder) SmsLogin(ctx, smsLoginReq interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SmsLogin", reflect.TypeOf((*MockIUserLogic)(nil).SmsLogin), ctx, smsLoginReq)
}