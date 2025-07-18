// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: kessel/inventory/v1beta2/streamed_list_objects_request.proto

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

type StreamedListObjectsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ObjectType    *RepresentationType    `protobuf:"bytes,1,opt,name=object_type,json=objectType,proto3" json:"object_type,omitempty"`
	Relation      string                 `protobuf:"bytes,2,opt,name=relation,proto3" json:"relation,omitempty"`
	Subject       *SubjectReference      `protobuf:"bytes,3,opt,name=subject,proto3" json:"subject,omitempty"`
	Pagination    *RequestPagination     `protobuf:"bytes,4,opt,name=pagination,proto3,oneof" json:"pagination,omitempty"`
	Consistency   *Consistency           `protobuf:"bytes,5,opt,name=consistency,proto3,oneof" json:"consistency,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamedListObjectsRequest) Reset() {
	*x = StreamedListObjectsRequest{}
	mi := &file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamedListObjectsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamedListObjectsRequest) ProtoMessage() {}

func (x *StreamedListObjectsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamedListObjectsRequest.ProtoReflect.Descriptor instead.
func (*StreamedListObjectsRequest) Descriptor() ([]byte, []int) {
	return file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescGZIP(), []int{0}
}

func (x *StreamedListObjectsRequest) GetObjectType() *RepresentationType {
	if x != nil {
		return x.ObjectType
	}
	return nil
}

func (x *StreamedListObjectsRequest) GetRelation() string {
	if x != nil {
		return x.Relation
	}
	return ""
}

func (x *StreamedListObjectsRequest) GetSubject() *SubjectReference {
	if x != nil {
		return x.Subject
	}
	return nil
}

func (x *StreamedListObjectsRequest) GetPagination() *RequestPagination {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *StreamedListObjectsRequest) GetConsistency() *Consistency {
	if x != nil {
		return x.Consistency
	}
	return nil
}

var File_kessel_inventory_v1beta2_streamed_list_objects_request_proto protoreflect.FileDescriptor

const file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDesc = "" +
	"\n" +
	"<kessel/inventory/v1beta2/streamed_list_objects_request.proto\x12\x18kessel.inventory.v1beta2\x1a\x1bbuf/validate/validate.proto\x1a1kessel/inventory/v1beta2/request_pagination.proto\x1a0kessel/inventory/v1beta2/subject_reference.proto\x1a*kessel/inventory/v1beta2/consistency.proto\x1a2kessel/inventory/v1beta2/representation_type.proto\"\xa5\x03\n" +
	"\x1aStreamedListObjectsRequest\x12U\n" +
	"\vobject_type\x18\x01 \x01(\v2,.kessel.inventory.v1beta2.RepresentationTypeB\x06\xbaH\x03\xc8\x01\x01R\n" +
	"objectType\x12#\n" +
	"\brelation\x18\x02 \x01(\tB\a\xbaH\x04r\x02\x10\x01R\brelation\x12L\n" +
	"\asubject\x18\x03 \x01(\v2*.kessel.inventory.v1beta2.SubjectReferenceB\x06\xbaH\x03\xc8\x01\x01R\asubject\x12P\n" +
	"\n" +
	"pagination\x18\x04 \x01(\v2+.kessel.inventory.v1beta2.RequestPaginationH\x00R\n" +
	"pagination\x88\x01\x01\x12L\n" +
	"\vconsistency\x18\x05 \x01(\v2%.kessel.inventory.v1beta2.ConsistencyH\x01R\vconsistency\x88\x01\x01B\r\n" +
	"\v_paginationB\x0e\n" +
	"\f_consistencyBr\n" +
	"(org.project_kessel.api.inventory.v1beta2P\x01ZDgithub.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2b\x06proto3"

var (
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescOnce sync.Once
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescData []byte
)

func file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescGZIP() []byte {
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescOnce.Do(func() {
		file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDesc), len(file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDesc)))
	})
	return file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDescData
}

var file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_goTypes = []any{
	(*StreamedListObjectsRequest)(nil), // 0: kessel.inventory.v1beta2.StreamedListObjectsRequest
	(*RepresentationType)(nil),         // 1: kessel.inventory.v1beta2.RepresentationType
	(*SubjectReference)(nil),           // 2: kessel.inventory.v1beta2.SubjectReference
	(*RequestPagination)(nil),          // 3: kessel.inventory.v1beta2.RequestPagination
	(*Consistency)(nil),                // 4: kessel.inventory.v1beta2.Consistency
}
var file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_depIdxs = []int32{
	1, // 0: kessel.inventory.v1beta2.StreamedListObjectsRequest.object_type:type_name -> kessel.inventory.v1beta2.RepresentationType
	2, // 1: kessel.inventory.v1beta2.StreamedListObjectsRequest.subject:type_name -> kessel.inventory.v1beta2.SubjectReference
	3, // 2: kessel.inventory.v1beta2.StreamedListObjectsRequest.pagination:type_name -> kessel.inventory.v1beta2.RequestPagination
	4, // 3: kessel.inventory.v1beta2.StreamedListObjectsRequest.consistency:type_name -> kessel.inventory.v1beta2.Consistency
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_init() }
func file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_init() {
	if File_kessel_inventory_v1beta2_streamed_list_objects_request_proto != nil {
		return
	}
	file_kessel_inventory_v1beta2_request_pagination_proto_init()
	file_kessel_inventory_v1beta2_subject_reference_proto_init()
	file_kessel_inventory_v1beta2_consistency_proto_init()
	file_kessel_inventory_v1beta2_representation_type_proto_init()
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDesc), len(file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_goTypes,
		DependencyIndexes: file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_depIdxs,
		MessageInfos:      file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_msgTypes,
	}.Build()
	File_kessel_inventory_v1beta2_streamed_list_objects_request_proto = out.File
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_goTypes = nil
	file_kessel_inventory_v1beta2_streamed_list_objects_request_proto_depIdxs = nil
}
