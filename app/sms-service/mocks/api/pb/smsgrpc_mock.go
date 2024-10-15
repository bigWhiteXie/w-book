// Code generated by MockGen. DO NOT EDIT.
// Source: api/pb/sms_grpc.pb.go

// Package mock_pb is a generated GoMock package.
package mock_pb

import (
	context "context"
	reflect "reflect"

	pb "codexie.com/w-book-code/api/pb"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockCodeClient is a mock of CodeClient interface.
type MockCodeClient struct {
	ctrl     *gomock.Controller
	recorder *MockCodeClientMockRecorder
}

// MockCodeClientMockRecorder is the mock recorder for MockCodeClient.
type MockCodeClientMockRecorder struct {
	mock *MockCodeClient
}

// NewMockCodeClient creates a new mock instance.
func NewMockCodeClient(ctrl *gomock.Controller) *MockCodeClient {
	mock := &MockCodeClient{ctrl: ctrl}
	mock.recorder = &MockCodeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeClient) EXPECT() *MockCodeClientMockRecorder {
	return m.recorder
}

// SendCode mocks base method.
func (m *MockCodeClient) SendCode(ctx context.Context, in *pb.SendCodeReq, opts ...grpc.CallOption) (*pb.SendCodeResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SendCode", varargs...)
	ret0, _ := ret[0].(*pb.SendCodeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendCode indicates an expected call of SendCode.
func (mr *MockCodeClientMockRecorder) SendCode(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCode", reflect.TypeOf((*MockCodeClient)(nil).SendCode), varargs...)
}

// VerifyCode mocks base method.
func (m *MockCodeClient) VerifyCode(ctx context.Context, in *pb.VerifyCodeReq, opts ...grpc.CallOption) (*pb.VerifyCodeResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "VerifyCode", varargs...)
	ret0, _ := ret[0].(*pb.VerifyCodeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyCode indicates an expected call of VerifyCode.
func (mr *MockCodeClientMockRecorder) VerifyCode(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyCode", reflect.TypeOf((*MockCodeClient)(nil).VerifyCode), varargs...)
}

// MockSMSServer is a mock of SMSServer interface.
type MockSMSServer struct {
	ctrl     *gomock.Controller
	recorder *MockSMSServerMockRecorder
}

// MockSMSServerMockRecorder is the mock recorder for MockSMSServer.
type MockSMSServerMockRecorder struct {
	mock *MockSMSServer
}

// NewMockSMSServer creates a new mock instance.
func NewMockSMSServer(ctrl *gomock.Controller) *MockSMSServer {
	mock := &MockSMSServer{ctrl: ctrl}
	mock.recorder = &MockSMSServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSMSServer) EXPECT() *MockSMSServerMockRecorder {
	return m.recorder
}

// SendCode mocks base method.
func (m *MockSMSServer) SendCode(arg0 context.Context, arg1 *pb.SendCodeReq) (*pb.SendCodeResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCode", arg0, arg1)
	ret0, _ := ret[0].(*pb.SendCodeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendCode indicates an expected call of SendCode.
func (mr *MockSMSServerMockRecorder) SendCode(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCode", reflect.TypeOf((*MockSMSServer)(nil).SendCode), arg0, arg1)
}

// VerifyCode mocks base method.
func (m *MockSMSServer) VerifyCode(arg0 context.Context, arg1 *pb.VerifyCodeReq) (*pb.VerifyCodeResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyCode", arg0, arg1)
	ret0, _ := ret[0].(*pb.VerifyCodeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyCode indicates an expected call of VerifyCode.
func (mr *MockSMSServerMockRecorder) VerifyCode(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyCode", reflect.TypeOf((*MockSMSServer)(nil).VerifyCode), arg0, arg1)
}

// mustEmbedUnimplementedSMSServer mocks base method.
func (m *MockSMSServer) mustEmbedUnimplementedSMSServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedSMSServer")
}

// mustEmbedUnimplementedSMSServer indicates an expected call of mustEmbedUnimplementedSMSServer.
func (mr *MockSMSServerMockRecorder) mustEmbedUnimplementedSMSServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedSMSServer", reflect.TypeOf((*MockSMSServer)(nil).mustEmbedUnimplementedSMSServer))
}

// MockUnsafeSMSServer is a mock of UnsafeSMSServer interface.
type MockUnsafeSMSServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeSMSServerMockRecorder
}

// MockUnsafeSMSServerMockRecorder is the mock recorder for MockUnsafeSMSServer.
type MockUnsafeSMSServerMockRecorder struct {
	mock *MockUnsafeSMSServer
}

// NewMockUnsafeSMSServer creates a new mock instance.
func NewMockUnsafeSMSServer(ctrl *gomock.Controller) *MockUnsafeSMSServer {
	mock := &MockUnsafeSMSServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeSMSServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeSMSServer) EXPECT() *MockUnsafeSMSServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedSMSServer mocks base method.
func (m *MockUnsafeSMSServer) mustEmbedUnimplementedSMSServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedSMSServer")
}

// mustEmbedUnimplementedSMSServer indicates an expected call of mustEmbedUnimplementedSMSServer.
func (mr *MockUnsafeSMSServerMockRecorder) mustEmbedUnimplementedSMSServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedSMSServer", reflect.TypeOf((*MockUnsafeSMSServer)(nil).mustEmbedUnimplementedSMSServer))
}