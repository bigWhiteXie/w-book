// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.1
// source: interact.proto

// proto 包名

package interact

import (
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

// 查询交互计数请求
type QueryInteractionReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=Biz,proto3" json:"Biz,omitempty"`      //业务类型
	BizId int64  `protobuf:"varint,2,opt,name=BizId,proto3" json:"BizId,omitempty"` //id
}

func (x *QueryInteractionReq) Reset() {
	*x = QueryInteractionReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_interact_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryInteractionReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryInteractionReq) ProtoMessage() {}

func (x *QueryInteractionReq) ProtoReflect() protoreflect.Message {
	mi := &file_interact_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryInteractionReq.ProtoReflect.Descriptor instead.
func (*QueryInteractionReq) Descriptor() ([]byte, []int) {
	return file_interact_proto_rawDescGZIP(), []int{0}
}

func (x *QueryInteractionReq) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *QueryInteractionReq) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

// 增加阅读量请求
type AddReadCntReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=Biz,proto3" json:"Biz,omitempty"`      //业务类型
	BizId int64  `protobuf:"varint,2,opt,name=BizId,proto3" json:"BizId,omitempty"` //id
}

func (x *AddReadCntReq) Reset() {
	*x = AddReadCntReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_interact_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddReadCntReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddReadCntReq) ProtoMessage() {}

func (x *AddReadCntReq) ProtoReflect() protoreflect.Message {
	mi := &file_interact_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddReadCntReq.ProtoReflect.Descriptor instead.
func (*AddReadCntReq) Descriptor() ([]byte, []int) {
	return file_interact_proto_rawDescGZIP(), []int{1}
}

func (x *AddReadCntReq) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *AddReadCntReq) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

// 通用响应体
type InteractionResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReadCnt    int64 `protobuf:"varint,1,opt,name=ReadCnt,proto3" json:"ReadCnt,omitempty"`
	CollectCnt int64 `protobuf:"varint,2,opt,name=CollectCnt,proto3" json:"CollectCnt,omitempty"`
	LikeCnt    int64 `protobuf:"varint,3,opt,name=LikeCnt,proto3" json:"LikeCnt,omitempty"`
}

func (x *InteractionResult) Reset() {
	*x = InteractionResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_interact_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InteractionResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InteractionResult) ProtoMessage() {}

func (x *InteractionResult) ProtoReflect() protoreflect.Message {
	mi := &file_interact_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InteractionResult.ProtoReflect.Descriptor instead.
func (*InteractionResult) Descriptor() ([]byte, []int) {
	return file_interact_proto_rawDescGZIP(), []int{2}
}

func (x *InteractionResult) GetReadCnt() int64 {
	if x != nil {
		return x.ReadCnt
	}
	return 0
}

func (x *InteractionResult) GetCollectCnt() int64 {
	if x != nil {
		return x.CollectCnt
	}
	return 0
}

func (x *InteractionResult) GetLikeCnt() int64 {
	if x != nil {
		return x.LikeCnt
	}
	return 0
}

type CommonResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg string `protobuf:"bytes,1,opt,name=Msg,proto3" json:"Msg,omitempty"`
}

func (x *CommonResult) Reset() {
	*x = CommonResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_interact_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommonResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommonResult) ProtoMessage() {}

func (x *CommonResult) ProtoReflect() protoreflect.Message {
	mi := &file_interact_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommonResult.ProtoReflect.Descriptor instead.
func (*CommonResult) Descriptor() ([]byte, []int) {
	return file_interact_proto_rawDescGZIP(), []int{3}
}

func (x *CommonResult) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

var File_interact_proto protoreflect.FileDescriptor

var file_interact_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x03, 0x61, 0x70, 0x69, 0x22, 0x3d, 0x0a, 0x13, 0x51, 0x75, 0x65, 0x72, 0x79, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03,
	0x42, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x42, 0x69, 0x7a, 0x12, 0x14,
	0x0a, 0x05, 0x42, 0x69, 0x7a, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x42,
	0x69, 0x7a, 0x49, 0x64, 0x22, 0x37, 0x0a, 0x0d, 0x41, 0x64, 0x64, 0x52, 0x65, 0x61, 0x64, 0x43,
	0x6e, 0x74, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x42, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x42, 0x69, 0x7a, 0x12, 0x14, 0x0a, 0x05, 0x42, 0x69, 0x7a, 0x49, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x42, 0x69, 0x7a, 0x49, 0x64, 0x22, 0x67, 0x0a,
	0x11, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x07, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x43, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0a, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x43, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x4c, 0x69, 0x6b, 0x65, 0x43, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x4c,
	0x69, 0x6b, 0x65, 0x43, 0x6e, 0x74, 0x22, 0x20, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x4d, 0x73, 0x67, 0x32, 0x8e, 0x01, 0x0a, 0x0b, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x48, 0x0a, 0x14, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x18, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x49, 0x6e, 0x74, 0x65,
	0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x16, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x35, 0x0a, 0x0c, 0x49, 0x6e, 0x63, 0x72, 0x65, 0x52, 0x65, 0x61, 0x64, 0x43,
	0x6e, 0x74, 0x12, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x41, 0x64, 0x64, 0x52, 0x65, 0x61, 0x64,
	0x43, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x11, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x0d, 0x5a, 0x0b, 0x70, 0x62, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_interact_proto_rawDescOnce sync.Once
	file_interact_proto_rawDescData = file_interact_proto_rawDesc
)

func file_interact_proto_rawDescGZIP() []byte {
	file_interact_proto_rawDescOnce.Do(func() {
		file_interact_proto_rawDescData = protoimpl.X.CompressGZIP(file_interact_proto_rawDescData)
	})
	return file_interact_proto_rawDescData
}

var file_interact_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_interact_proto_goTypes = []any{
	(*QueryInteractionReq)(nil), // 0: api.QueryInteractionReq
	(*AddReadCntReq)(nil),       // 1: api.AddReadCntReq
	(*InteractionResult)(nil),   // 2: api.InteractionResult
	(*CommonResult)(nil),        // 3: api.CommonResult
}
var file_interact_proto_depIdxs = []int32{
	0, // 0: api.Interaction.QueryInteractionInfo:input_type -> api.QueryInteractionReq
	1, // 1: api.Interaction.IncreReadCnt:input_type -> api.AddReadCntReq
	2, // 2: api.Interaction.QueryInteractionInfo:output_type -> api.InteractionResult
	3, // 3: api.Interaction.IncreReadCnt:output_type -> api.CommonResult
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_interact_proto_init() }
func file_interact_proto_init() {
	if File_interact_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_interact_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*QueryInteractionReq); i {
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
		file_interact_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*AddReadCntReq); i {
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
		file_interact_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*InteractionResult); i {
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
		file_interact_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*CommonResult); i {
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
			RawDescriptor: file_interact_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_interact_proto_goTypes,
		DependencyIndexes: file_interact_proto_depIdxs,
		MessageInfos:      file_interact_proto_msgTypes,
	}.Build()
	File_interact_proto = out.File
	file_interact_proto_rawDesc = nil
	file_interact_proto_goTypes = nil
	file_interact_proto_depIdxs = nil
}