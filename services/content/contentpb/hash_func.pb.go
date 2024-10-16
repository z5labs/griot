// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.30.0--dev
// source: hash_func.proto

package contentpb

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

type HashFunc int32

const (
	HashFunc_SHA256 HashFunc = 0
)

// Enum value maps for HashFunc.
var (
	HashFunc_name = map[int32]string{
		0: "SHA256",
	}
	HashFunc_value = map[string]int32{
		"SHA256": 0,
	}
)

func (x HashFunc) Enum() *HashFunc {
	p := new(HashFunc)
	*p = x
	return p
}

func (x HashFunc) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HashFunc) Descriptor() protoreflect.EnumDescriptor {
	return file_hash_func_proto_enumTypes[0].Descriptor()
}

func (HashFunc) Type() protoreflect.EnumType {
	return &file_hash_func_proto_enumTypes[0]
}

func (x HashFunc) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HashFunc.Descriptor instead.
func (HashFunc) EnumDescriptor() ([]byte, []int) {
	return file_hash_func_proto_rawDescGZIP(), []int{0}
}

var File_hash_func_proto protoreflect.FileDescriptor

var file_hash_func_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x68, 0x61, 0x73, 0x68, 0x5f, 0x66, 0x75, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0d, 0x67, 0x72, 0x69, 0x6f, 0x74, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x2a, 0x16, 0x0a, 0x08, 0x48, 0x61, 0x73, 0x68, 0x46, 0x75, 0x6e, 0x63, 0x12, 0x0a, 0x0a, 0x06,
	0x53, 0x48, 0x41, 0x32, 0x35, 0x36, 0x10, 0x00, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x7a, 0x35, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x67, 0x72,
	0x69, 0x6f, 0x74, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x3b, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x62, 0x08, 0x65, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x70, 0xe8, 0x07,
}

var (
	file_hash_func_proto_rawDescOnce sync.Once
	file_hash_func_proto_rawDescData = file_hash_func_proto_rawDesc
)

func file_hash_func_proto_rawDescGZIP() []byte {
	file_hash_func_proto_rawDescOnce.Do(func() {
		file_hash_func_proto_rawDescData = protoimpl.X.CompressGZIP(file_hash_func_proto_rawDescData)
	})
	return file_hash_func_proto_rawDescData
}

var file_hash_func_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_hash_func_proto_goTypes = []any{
	(HashFunc)(0), // 0: griot.content.HashFunc
}
var file_hash_func_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_hash_func_proto_init() }
func file_hash_func_proto_init() {
	if File_hash_func_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_hash_func_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_hash_func_proto_goTypes,
		DependencyIndexes: file_hash_func_proto_depIdxs,
		EnumInfos:         file_hash_func_proto_enumTypes,
	}.Build()
	File_hash_func_proto = out.File
	file_hash_func_proto_rawDesc = nil
	file_hash_func_proto_goTypes = nil
	file_hash_func_proto_depIdxs = nil
}
