// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.3
// source: microserviceCommunication.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MicroserviceCommunication struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type        string            `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Data        *structpb.Struct  `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Metadata    map[string]string `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	UserRequest *anypb.Any        `protobuf:"bytes,4,opt,name=user_request,json=userRequest,proto3" json:"user_request,omitempty"`
}

func (x *MicroserviceCommunication) Reset() {
	*x = MicroserviceCommunication{}
	if protoimpl.UnsafeEnabled {
		mi := &file_microserviceCommunication_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MicroserviceCommunication) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MicroserviceCommunication) ProtoMessage() {}

func (x *MicroserviceCommunication) ProtoReflect() protoreflect.Message {
	mi := &file_microserviceCommunication_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MicroserviceCommunication.ProtoReflect.Descriptor instead.
func (*MicroserviceCommunication) Descriptor() ([]byte, []int) {
	return file_microserviceCommunication_proto_rawDescGZIP(), []int{0}
}

func (x *MicroserviceCommunication) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *MicroserviceCommunication) GetData() *structpb.Struct {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *MicroserviceCommunication) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *MicroserviceCommunication) GetUserRequest() *anypb.Any {
	if x != nil {
		return x.UserRequest
	}
	return nil
}

var File_microserviceCommunication_proto protoreflect.FileDescriptor

var file_microserviceCommunication_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f,
	0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9e,
	0x02, 0x0a, 0x19, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43,
	0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x2b, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x4a, 0x0a,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x2e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x37, 0x0a, 0x0c, 0x75, 0x73, 0x65,
	0x72, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0b, 0x75, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32,
	0x56, 0x0a, 0x0c, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x46, 0x0a, 0x08, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x20, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x43, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x16, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4a, 0x6f, 0x72, 0x72, 0x69, 0x74, 0x30, 0x35, 0x2f, 0x44,
	0x59, 0x4e, 0x41, 0x4d, 0x4f, 0x53, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_microserviceCommunication_proto_rawDescOnce sync.Once
	file_microserviceCommunication_proto_rawDescData = file_microserviceCommunication_proto_rawDesc
)

func file_microserviceCommunication_proto_rawDescGZIP() []byte {
	file_microserviceCommunication_proto_rawDescOnce.Do(func() {
		file_microserviceCommunication_proto_rawDescData = protoimpl.X.CompressGZIP(file_microserviceCommunication_proto_rawDescData)
	})
	return file_microserviceCommunication_proto_rawDescData
}

var file_microserviceCommunication_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_microserviceCommunication_proto_goTypes = []interface{}{
	(*MicroserviceCommunication)(nil), // 0: proto.MicroserviceCommunication
	nil,                               // 1: proto.MicroserviceCommunication.MetadataEntry
	(*structpb.Struct)(nil),           // 2: google.protobuf.Struct
	(*anypb.Any)(nil),                 // 3: google.protobuf.Any
	(*emptypb.Empty)(nil),             // 4: google.protobuf.Empty
}
var file_microserviceCommunication_proto_depIdxs = []int32{
	2, // 0: proto.MicroserviceCommunication.data:type_name -> google.protobuf.Struct
	1, // 1: proto.MicroserviceCommunication.metadata:type_name -> proto.MicroserviceCommunication.MetadataEntry
	3, // 2: proto.MicroserviceCommunication.user_request:type_name -> google.protobuf.Any
	0, // 3: proto.Microservice.SendData:input_type -> proto.MicroserviceCommunication
	4, // 4: proto.Microservice.SendData:output_type -> google.protobuf.Empty
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_microserviceCommunication_proto_init() }
func file_microserviceCommunication_proto_init() {
	if File_microserviceCommunication_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_microserviceCommunication_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MicroserviceCommunication); i {
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
			RawDescriptor: file_microserviceCommunication_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_microserviceCommunication_proto_goTypes,
		DependencyIndexes: file_microserviceCommunication_proto_depIdxs,
		MessageInfos:      file_microserviceCommunication_proto_msgTypes,
	}.Build()
	File_microserviceCommunication_proto = out.File
	file_microserviceCommunication_proto_rawDesc = nil
	file_microserviceCommunication_proto_goTypes = nil
	file_microserviceCommunication_proto_depIdxs = nil
}
