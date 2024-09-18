// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.1
// source: err.proto

package codeerr

import (
	"codexie.com/w-book-user/pkg/common/response"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WithCodeErr struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code int32  `protobuf:"varint,1,opt,name=Code,proto3" json:"Code,omitempty"`
	Msg  string `protobuf:"bytes,2,opt,name=Msg,proto3" json:"Msg,omitempty"`
}

func WithCode(code int, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.Wrap(&WithCodeErr{
		Msg:  msg,
		Code: int32(code),
	}, msg)
}

func ParseGrpcErr(err error) *WithCodeErr {
	st, ok := status.FromError(err)
	if !ok {
		return &WithCodeErr{Code: CodeSystemERR, Msg: err.Error()}
	}
	switch st.Code() {
	case codes.Internal:
		for _, d := range st.Details() {
			switch d.(type) {
			case *WithCodeErr:
				e := d.(*WithCodeErr)
				return e
			}

		}
	}
	return &WithCodeErr{Code: SystemErrCpde, Msg: st.Message()}
}

func (e *WithCodeErr) StackTrace() errors.StackTrace {
	// Get full stack trace
	stack := errors.WithStack(e).(interface {
		StackTrace() errors.StackTrace
	}).StackTrace()

	// Remove the top frame (which corresponds to the `WithCodeErrErr` call)
	return stack[1:] // Skip the first stack frame
}

func ToGrpcErr(e error) error {
	st := status.New(codes.Internal, e.Error())
	codeErr := &WithCodeErr{}
	if ok := errors.As(e, &codeErr); !ok {
		return st.Err()
	}
	details, err := st.WithDetails(codeErr)
	if err == nil {
		return details.Err()
	}
	return st.Err()
}

// HandleErr 日志记录异常并封装返回给用户的异常响应
//
//	@Description:
//	@param ctx
//	@param err
//	@return *response.Response
func HandleErr(ctx context.Context, err error) *response.Response {
	logx.WithContext(ctx).Errorf("%+v", err)
	coder := ParseCoder(err)
	return response.Fail(coder.Code(), coder.String())
}

func (w *WithCodeErr) Error() string { return w.Msg }
func (x *WithCodeErr) Reset() {
	*x = WithCodeErr{}
	if protoimpl.UnsafeEnabled {
		mi := &file_err_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WithCodeErr) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WithCodeErr) ProtoMessage() {}

func (x *WithCodeErr) ProtoReflect() protoreflect.Message {
	mi := &file_err_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WithCodeErr.ProtoReflect.Descriptor instead.
func (*WithCodeErr) Descriptor() ([]byte, []int) {
	return file_err_proto_rawDescGZIP(), []int{0}
}

func (x *WithCodeErr) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *WithCodeErr) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

var File_err_proto protoreflect.FileDescriptor

var file_err_proto_rawDesc = []byte{
	0x0a, 0x09, 0x65, 0x72, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70, 0x69,
	0x22, 0x33, 0x0a, 0x0b, 0x57, 0x69, 0x74, 0x68, 0x43, 0x6f, 0x64, 0x65, 0x45, 0x72, 0x72, 0x12,
	0x12, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x43,
	0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x4d, 0x73, 0x67, 0x42, 0x19, 0x5a, 0x17, 0x63, 0x6f, 0x64, 0x65, 0x78, 0x69, 0x65,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x2d, 0x62, 0x6f, 0x6f, 0x6b, 0x2d, 0x75, 0x73, 0x65, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_err_proto_rawDescOnce sync.Once
	file_err_proto_rawDescData = file_err_proto_rawDesc
)

func file_err_proto_rawDescGZIP() []byte {
	file_err_proto_rawDescOnce.Do(func() {
		file_err_proto_rawDescData = protoimpl.X.CompressGZIP(file_err_proto_rawDescData)
	})
	return file_err_proto_rawDescData
}

var file_err_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_err_proto_goTypes = []any{
	(*WithCodeErr)(nil), // 0: api.WithCodeErr
}
var file_err_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_err_proto_init() }
func file_err_proto_init() {
	if File_err_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_err_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*WithCodeErr); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_err_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_err_proto_goTypes,
		DependencyIndexes: file_err_proto_depIdxs,
		MessageInfos:      file_err_proto_msgTypes,
	}.Build()
	File_err_proto = out.File
	file_err_proto_rawDesc = nil
	file_err_proto_goTypes = nil
	file_err_proto_depIdxs = nil
}