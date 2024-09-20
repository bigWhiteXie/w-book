// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.1
// source: sms.proto

// proto 包名

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	SMS_SendCode_FullMethodName   = "/api.SMS/SendCode"
	SMS_VerifyCode_FullMethodName = "/api.SMS/VerifyCode"
)

// CodeClient is the client API for SMS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// mockgen -source=api/pb/sms_grpc.pb.go -destination mocks/api/pb/smsgrpc_mock.go
type CodeClient interface {
	SendCode(ctx context.Context, in *SendCodeReq, opts ...grpc.CallOption) (*SendCodeResp, error)
	VerifyCode(ctx context.Context, in *VerifyCodeReq, opts ...grpc.CallOption) (*VerifyCodeResp, error)
}

type codeClient struct {
	cc grpc.ClientConnInterface
}

func NewCodeClient(cc grpc.ClientConnInterface) CodeClient {
	return &codeClient{cc}
}

func (c *codeClient) SendCode(ctx context.Context, in *SendCodeReq, opts ...grpc.CallOption) (*SendCodeResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendCodeResp)
	err := c.cc.Invoke(ctx, SMS_SendCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *codeClient) VerifyCode(ctx context.Context, in *VerifyCodeReq, opts ...grpc.CallOption) (*VerifyCodeResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyCodeResp)
	err := c.cc.Invoke(ctx, SMS_VerifyCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SMSServer is the server API for SMS service.
// All implementations must embed UnimplementedSMSServer
// for forward compatibility.
//
// 定义 Greet 服务
type SMSServer interface {
	SendCode(context.Context, *SendCodeReq) (*SendCodeResp, error)
	VerifyCode(context.Context, *VerifyCodeReq) (*VerifyCodeResp, error)
	mustEmbedUnimplementedSMSServer()
}

// UnimplementedSMSServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSMSServer struct{}

func (UnimplementedSMSServer) SendCode(context.Context, *SendCodeReq) (*SendCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCode not implemented")
}
func (UnimplementedSMSServer) VerifyCode(context.Context, *VerifyCodeReq) (*VerifyCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyCode not implemented")
}
func (UnimplementedSMSServer) mustEmbedUnimplementedSMSServer() {}
func (UnimplementedSMSServer) testEmbeddedByValue()             {}

// UnsafeSMSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SMSServer will
// result in compilation errors.
type UnsafeSMSServer interface {
	mustEmbedUnimplementedSMSServer()
}

func RegisterSMSServer(s grpc.ServiceRegistrar, srv SMSServer) {
	// If the following call pancis, it indicates UnimplementedSMSServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SMS_ServiceDesc, srv)
}

func _SMS_SendCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SMSServer).SendCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SMS_SendCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SMSServer).SendCode(ctx, req.(*SendCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SMS_VerifyCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SMSServer).VerifyCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SMS_VerifyCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SMSServer).VerifyCode(ctx, req.(*VerifyCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

// SMS_ServiceDesc is the grpc.ServiceDesc for SMS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SMS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.SMS",
	HandlerType: (*SMSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendCode",
			Handler:    _SMS_SendCode_Handler,
		},
		{
			MethodName: "VerifyCode",
			Handler:    _SMS_VerifyCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sms.proto",
}
