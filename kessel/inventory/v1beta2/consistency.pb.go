// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: kessel/inventory/v1beta2/consistency.proto

package v1beta2

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Defines how a request is handled by the service.
type Consistency struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Requirement:
	//
	//	*Consistency_MinimizeLatency
	//	*Consistency_AtLeastAsFresh
	Requirement   isConsistency_Requirement `protobuf_oneof:"requirement"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Consistency) Reset() {
	*x = Consistency{}
	mi := &file_kessel_inventory_v1beta2_consistency_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Consistency) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Consistency) ProtoMessage() {}

func (x *Consistency) ProtoReflect() protoreflect.Message {
	mi := &file_kessel_inventory_v1beta2_consistency_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Consistency.ProtoReflect.Descriptor instead.
func (*Consistency) Descriptor() ([]byte, []int) {
	return file_kessel_inventory_v1beta2_consistency_proto_rawDescGZIP(), []int{0}
}

func (x *Consistency) GetRequirement() isConsistency_Requirement {
	if x != nil {
		return x.Requirement
	}
	return nil
}

func (x *Consistency) GetMinimizeLatency() bool {
	if x != nil {
		if x, ok := x.Requirement.(*Consistency_MinimizeLatency); ok {
			return x.MinimizeLatency
		}
	}
	return false
}

func (x *Consistency) GetAtLeastAsFresh() *ConsistencyToken {
	if x != nil {
		if x, ok := x.Requirement.(*Consistency_AtLeastAsFresh); ok {
			return x.AtLeastAsFresh
		}
	}
	return nil
}

type isConsistency_Requirement interface {
	isConsistency_Requirement()
}

type Consistency_MinimizeLatency struct {
	// The service selects the fastest snapshot available.
	// *Must* be set true if used.
	MinimizeLatency bool `protobuf:"varint,1,opt,name=minimize_latency,json=minimizeLatency,proto3,oneof"`
}

type Consistency_AtLeastAsFresh struct {
	// All data used in the API call must be *at least as fresh*
	// as found in the ConsistencyToken. More recent data might be used
	// if available or faster.
	AtLeastAsFresh *ConsistencyToken `protobuf:"bytes,2,opt,name=at_least_as_fresh,json=atLeastAsFresh,proto3,oneof"`
}

func (*Consistency_MinimizeLatency) isConsistency_Requirement() {}

func (*Consistency_AtLeastAsFresh) isConsistency_Requirement() {}

var File_kessel_inventory_v1beta2_consistency_proto protoreflect.FileDescriptor

const file_kessel_inventory_v1beta2_consistency_proto_rawDesc = "" +
	"\n" +
	"*kessel/inventory/v1beta2/consistency.proto\x12\x18kessel.inventory.v1beta2\x1a\x1bbuf/validate/validate.proto\x1a0kessel/inventory/v1beta2/consistency_token.proto\"\xb2\x01\n" +
	"\vConsistency\x124\n" +
	"\x10minimize_latency\x18\x01 \x01(\bB\a\xbaH\x04j\x02\b\x01H\x00R\x0fminimizeLatency\x12W\n" +
	"\x11at_least_as_fresh\x18\x02 \x01(\v2*.kessel.inventory.v1beta2.ConsistencyTokenH\x00R\x0eatLeastAsFreshB\x14\n" +
	"\vrequirement\x12\x05\xbaH\x02\b\x01Br\n" +
	"(org.project_kessel.api.inventory.v1beta2P\x01ZDgithub.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2b\x06proto3"

var (
	file_kessel_inventory_v1beta2_consistency_proto_rawDescOnce sync.Once
	file_kessel_inventory_v1beta2_consistency_proto_rawDescData []byte
)

func file_kessel_inventory_v1beta2_consistency_proto_rawDescGZIP() []byte {
	file_kessel_inventory_v1beta2_consistency_proto_rawDescOnce.Do(func() {
		file_kessel_inventory_v1beta2_consistency_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_kessel_inventory_v1beta2_consistency_proto_rawDesc), len(file_kessel_inventory_v1beta2_consistency_proto_rawDesc)))
	})
	return file_kessel_inventory_v1beta2_consistency_proto_rawDescData
}

var file_kessel_inventory_v1beta2_consistency_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kessel_inventory_v1beta2_consistency_proto_goTypes = []any{
	(*Consistency)(nil),      // 0: kessel.inventory.v1beta2.Consistency
	(*ConsistencyToken)(nil), // 1: kessel.inventory.v1beta2.ConsistencyToken
}
var file_kessel_inventory_v1beta2_consistency_proto_depIdxs = []int32{
	1, // 0: kessel.inventory.v1beta2.Consistency.at_least_as_fresh:type_name -> kessel.inventory.v1beta2.ConsistencyToken
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kessel_inventory_v1beta2_consistency_proto_init() }
func file_kessel_inventory_v1beta2_consistency_proto_init() {
	if File_kessel_inventory_v1beta2_consistency_proto != nil {
		return
	}
	file_kessel_inventory_v1beta2_consistency_token_proto_init()
	file_kessel_inventory_v1beta2_consistency_proto_msgTypes[0].OneofWrappers = []any{
		(*Consistency_MinimizeLatency)(nil),
		(*Consistency_AtLeastAsFresh)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_kessel_inventory_v1beta2_consistency_proto_rawDesc), len(file_kessel_inventory_v1beta2_consistency_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kessel_inventory_v1beta2_consistency_proto_goTypes,
		DependencyIndexes: file_kessel_inventory_v1beta2_consistency_proto_depIdxs,
		MessageInfos:      file_kessel_inventory_v1beta2_consistency_proto_msgTypes,
	}.Build()
	File_kessel_inventory_v1beta2_consistency_proto = out.File
	file_kessel_inventory_v1beta2_consistency_proto_goTypes = nil
	file_kessel_inventory_v1beta2_consistency_proto_depIdxs = nil
}
