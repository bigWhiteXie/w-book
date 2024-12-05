// Code generated by MockGen. DO NOT EDIT.
// Source: interact_grpc.pb.go

// Package grpc is a generated GoMock package.
package grpc

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockInteractionClient is a mock of InteractionClient interface.
type MockInteractionClient struct {
	ctrl     *gomock.Controller
	recorder *MockInteractionClientMockRecorder
}

// MockInteractionClientMockRecorder is the mock recorder for MockInteractionClient.
type MockInteractionClientMockRecorder struct {
	mock *MockInteractionClient
}

// NewMockInteractionClient creates a new mock instance.
func NewMockInteractionClient(ctrl *gomock.Controller) *MockInteractionClient {
	mock := &MockInteractionClient{ctrl: ctrl}
	mock.recorder = &MockInteractionClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInteractionClient) EXPECT() *MockInteractionClientMockRecorder {
	return m.recorder
}

// IncreReadCnt mocks base method.
func (m *MockInteractionClient) IncreReadCnt(ctx context.Context, in *AddReadCntReq, opts ...grpc.CallOption) (*CommonResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "IncreReadCnt", varargs...)
	ret0, _ := ret[0].(*CommonResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IncreReadCnt indicates an expected call of IncreReadCnt.
func (mr *MockInteractionClientMockRecorder) IncreReadCnt(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncreReadCnt", reflect.TypeOf((*MockInteractionClient)(nil).IncreReadCnt), varargs...)
}

// QueryInteractionInfo mocks base method.
func (m *MockInteractionClient) QueryInteractionInfo(ctx context.Context, in *QueryInteractionReq, opts ...grpc.CallOption) (*InteractionResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryInteractionInfo", varargs...)
	ret0, _ := ret[0].(*InteractionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryInteractionInfo indicates an expected call of QueryInteractionInfo.
func (mr *MockInteractionClientMockRecorder) QueryInteractionInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryInteractionInfo", reflect.TypeOf((*MockInteractionClient)(nil).QueryInteractionInfo), varargs...)
}

// QueryInteractionsInfo mocks base method.
func (m *MockInteractionClient) QueryInteractionsInfo(ctx context.Context, in *QueryInteractionsReq, opts ...grpc.CallOption) (*InteractionsInfo, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryInteractionsInfo", varargs...)
	ret0, _ := ret[0].(*InteractionsInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryInteractionsInfo indicates an expected call of QueryInteractionsInfo.
func (mr *MockInteractionClientMockRecorder) QueryInteractionsInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryInteractionsInfo", reflect.TypeOf((*MockInteractionClient)(nil).QueryInteractionsInfo), varargs...)
}

// TopLike mocks base method.
func (m *MockInteractionClient) TopLike(ctx context.Context, in *TopLikeReq, opts ...grpc.CallOption) (*TopLikeResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "TopLike", varargs...)
	ret0, _ := ret[0].(*TopLikeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TopLike indicates an expected call of TopLike.
func (mr *MockInteractionClientMockRecorder) TopLike(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TopLike", reflect.TypeOf((*MockInteractionClient)(nil).TopLike), varargs...)
}

// MockInteractionServer is a mock of InteractionServer interface.
type MockInteractionServer struct {
	ctrl     *gomock.Controller
	recorder *MockInteractionServerMockRecorder
}

// MockInteractionServerMockRecorder is the mock recorder for MockInteractionServer.
type MockInteractionServerMockRecorder struct {
	mock *MockInteractionServer
}

// NewMockInteractionServer creates a new mock instance.
func NewMockInteractionServer(ctrl *gomock.Controller) *MockInteractionServer {
	mock := &MockInteractionServer{ctrl: ctrl}
	mock.recorder = &MockInteractionServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInteractionServer) EXPECT() *MockInteractionServerMockRecorder {
	return m.recorder
}

// IncreReadCnt mocks base method.
func (m *MockInteractionServer) IncreReadCnt(arg0 context.Context, arg1 *AddReadCntReq) (*CommonResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncreReadCnt", arg0, arg1)
	ret0, _ := ret[0].(*CommonResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IncreReadCnt indicates an expected call of IncreReadCnt.
func (mr *MockInteractionServerMockRecorder) IncreReadCnt(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncreReadCnt", reflect.TypeOf((*MockInteractionServer)(nil).IncreReadCnt), arg0, arg1)
}

// QueryInteractionInfo mocks base method.
func (m *MockInteractionServer) QueryInteractionInfo(arg0 context.Context, arg1 *QueryInteractionReq) (*InteractionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryInteractionInfo", arg0, arg1)
	ret0, _ := ret[0].(*InteractionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryInteractionInfo indicates an expected call of QueryInteractionInfo.
func (mr *MockInteractionServerMockRecorder) QueryInteractionInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryInteractionInfo", reflect.TypeOf((*MockInteractionServer)(nil).QueryInteractionInfo), arg0, arg1)
}

// QueryInteractionsInfo mocks base method.
func (m *MockInteractionServer) QueryInteractionsInfo(arg0 context.Context, arg1 *QueryInteractionsReq) (*InteractionsInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryInteractionsInfo", arg0, arg1)
	ret0, _ := ret[0].(*InteractionsInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryInteractionsInfo indicates an expected call of QueryInteractionsInfo.
func (mr *MockInteractionServerMockRecorder) QueryInteractionsInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryInteractionsInfo", reflect.TypeOf((*MockInteractionServer)(nil).QueryInteractionsInfo), arg0, arg1)
}

// TopLike mocks base method.
func (m *MockInteractionServer) TopLike(arg0 context.Context, arg1 *TopLikeReq) (*TopLikeResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TopLike", arg0, arg1)
	ret0, _ := ret[0].(*TopLikeResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TopLike indicates an expected call of TopLike.
func (mr *MockInteractionServerMockRecorder) TopLike(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TopLike", reflect.TypeOf((*MockInteractionServer)(nil).TopLike), arg0, arg1)
}

// mustEmbedUnimplementedInteractionServer mocks base method.
func (m *MockInteractionServer) mustEmbedUnimplementedInteractionServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedInteractionServer")
}

// mustEmbedUnimplementedInteractionServer indicates an expected call of mustEmbedUnimplementedInteractionServer.
func (mr *MockInteractionServerMockRecorder) mustEmbedUnimplementedInteractionServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedInteractionServer", reflect.TypeOf((*MockInteractionServer)(nil).mustEmbedUnimplementedInteractionServer))
}

// MockUnsafeInteractionServer is a mock of UnsafeInteractionServer interface.
type MockUnsafeInteractionServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeInteractionServerMockRecorder
}

// MockUnsafeInteractionServerMockRecorder is the mock recorder for MockUnsafeInteractionServer.
type MockUnsafeInteractionServerMockRecorder struct {
	mock *MockUnsafeInteractionServer
}

// NewMockUnsafeInteractionServer creates a new mock instance.
func NewMockUnsafeInteractionServer(ctrl *gomock.Controller) *MockUnsafeInteractionServer {
	mock := &MockUnsafeInteractionServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeInteractionServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeInteractionServer) EXPECT() *MockUnsafeInteractionServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedInteractionServer mocks base method.
func (m *MockUnsafeInteractionServer) mustEmbedUnimplementedInteractionServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedInteractionServer")
}

// mustEmbedUnimplementedInteractionServer indicates an expected call of mustEmbedUnimplementedInteractionServer.
func (mr *MockUnsafeInteractionServerMockRecorder) mustEmbedUnimplementedInteractionServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedInteractionServer", reflect.TypeOf((*MockUnsafeInteractionServer)(nil).mustEmbedUnimplementedInteractionServer))
}
