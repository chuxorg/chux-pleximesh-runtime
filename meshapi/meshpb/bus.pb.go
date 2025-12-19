package meshpb

import (
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
)

const _ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)

type AckStatus int32

const (
	AckStatus_ACK_STATUS_UNSPECIFIED AckStatus = 0
	AckStatus_ACK_STATUS_ACCEPTED    AckStatus = 1
	AckStatus_ACK_STATUS_REJECTED    AckStatus = 2
)

var (
	AckStatus_name = map[int32]string{
		0: "ACK_STATUS_UNSPECIFIED",
		1: "ACK_STATUS_ACCEPTED",
		2: "ACK_STATUS_REJECTED",
	}
	AckStatus_value = map[string]int32{
		"ACK_STATUS_UNSPECIFIED": 0,
		"ACK_STATUS_ACCEPTED":    1,
		"ACK_STATUS_REJECTED":    2,
	}
)

func (x AckStatus) Enum() *AckStatus {
	p := new(AckStatus)
	*p = x
	return p
}

func (x AckStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AckStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_mesh_v1_bus_proto_enumTypes[0].Descriptor()
}

func (AckStatus) Type() protoreflect.EnumType {
	return &file_mesh_v1_bus_proto_enumTypes[0]
}

func (x AckStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AckStatus.Descriptor instead.
func (AckStatus) EnumDescriptor() ([]byte, []int) {
	return file_mesh_v1_bus_proto_rawDescGZIP(), []int{0}
}

type BusIngressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Envelope *EventEnvelope `protobuf:"bytes,1,opt,name=envelope,proto3" json:"envelope,omitempty"`
}

func (x *BusIngressRequest) Reset() {
	*x = BusIngressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_bus_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BusIngressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BusIngressRequest) ProtoMessage() {}

func (x *BusIngressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_bus_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*BusIngressRequest) Descriptor() ([]byte, []int) {
	return file_mesh_v1_bus_proto_rawDescGZIP(), []int{0}
}

func (x *BusIngressRequest) GetEnvelope() *EventEnvelope {
	if x != nil {
		return x.Envelope
	}
	return nil
}

type BusIngressResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId string    `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	Status  AckStatus `protobuf:"varint,2,opt,name=status,proto3,enum=mesh.v1.AckStatus" json:"status,omitempty"`
	Message string    `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *BusIngressResponse) Reset() {
	*x = BusIngressResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_bus_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BusIngressResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BusIngressResponse) ProtoMessage() {}

func (x *BusIngressResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_bus_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*BusIngressResponse) Descriptor() ([]byte, []int) {
	return file_mesh_v1_bus_proto_rawDescGZIP(), []int{1}
}

func (x *BusIngressResponse) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *BusIngressResponse) GetStatus() AckStatus {
	if x != nil {
		return x.Status
	}
	return AckStatus_ACK_STATUS_UNSPECIFIED
}

func (x *BusIngressResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SubscriptionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domains    []string `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"`
	EventTypes []string `protobuf:"bytes,2,rep,name=event_types,json=eventTypes,proto3" json:"event_types,omitempty"`
	AgentId    string   `protobuf:"bytes,3,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
}

func (x *SubscriptionRequest) Reset() {
	*x = SubscriptionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_bus_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionRequest) ProtoMessage() {}

func (x *SubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_bus_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_mesh_v1_bus_proto_rawDescGZIP(), []int{2}
}

func (x *SubscriptionRequest) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

func (x *SubscriptionRequest) GetEventTypes() []string {
	if x != nil {
		return x.EventTypes
	}
	return nil
}

func (x *SubscriptionRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

var File_mesh_v1_bus_proto protoreflect.FileDescriptor

var file_mesh_v1_bus_proto_rawDescOnce sync.Once
var file_mesh_v1_bus_proto_rawDescData = buildMeshV1BusProtoRawDesc()

func buildMeshV1BusProtoRawDesc() []byte {
	fd := &descriptorpb.FileDescriptorProto{
		Syntax:     strPtr("proto3"),
		Name:       strPtr("mesh/v1/bus.proto"),
		Package:    strPtr("mesh.v1"),
		Dependency: []string{"mesh/v1/event.proto"},
		Options: &descriptorpb.FileOptions{
			GoPackage: strPtr("github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"),
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: strPtr("AckStatus"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: strPtr("ACK_STATUS_UNSPECIFIED"), Number: int32Ptr(0)},
					{Name: strPtr("ACK_STATUS_ACCEPTED"), Number: int32Ptr(1)},
					{Name: strPtr("ACK_STATUS_REJECTED"), Number: int32Ptr(2)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: strPtr("BusIngressRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("envelope"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: strPtr(".mesh.v1.EventEnvelope"),
						JsonName: strPtr("envelope"),
					},
				},
			},
			{
				Name: strPtr("BusIngressResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("event_id"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("eventId"),
					},
					{
						Name:     strPtr("status"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: strPtr(".mesh.v1.AckStatus"),
						JsonName: strPtr("status"),
					},
					{
						Name:     strPtr("message"),
						Number:   int32Ptr(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("message"),
					},
				},
			},
			{
				Name: strPtr("SubscriptionRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("domains"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("domains"),
					},
					{
						Name:     strPtr("event_types"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("eventTypes"),
					},
					{
						Name:     strPtr("agent_id"),
						Number:   int32Ptr(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("agentId"),
					},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: strPtr("EventBus"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:            strPtr("Publish"),
						InputType:       strPtr(".mesh.v1.BusIngressRequest"),
						OutputType:      strPtr(".mesh.v1.BusIngressResponse"),
						ClientStreaming: boolPtr(true),
						ServerStreaming: boolPtr(true),
					},
					{
						Name:            strPtr("Subscribe"),
						InputType:       strPtr(".mesh.v1.SubscriptionRequest"),
						OutputType:      strPtr(".mesh.v1.EventEnvelope"),
						ServerStreaming: boolPtr(true),
					},
				},
			},
		},
	}
	data, err := proto.Marshal(fd)
	if err != nil {
		panic(err)
	}
	return data
}

func file_mesh_v1_bus_proto_rawDescGZIP() []byte {
	file_mesh_v1_bus_proto_rawDescOnce.Do(func() {
		file_mesh_v1_bus_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1_bus_proto_rawDescData)
	})
	return file_mesh_v1_bus_proto_rawDescData
}

var file_mesh_v1_bus_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mesh_v1_bus_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_mesh_v1_bus_proto_goTypes = []interface{}{
	(AckStatus)(0),              // 0: mesh.v1.AckStatus
	(*BusIngressRequest)(nil),   // 1: mesh.v1.BusIngressRequest
	(*BusIngressResponse)(nil),  // 2: mesh.v1.BusIngressResponse
	(*SubscriptionRequest)(nil), // 3: mesh.v1.SubscriptionRequest
	(*EventEnvelope)(nil),       // 4: mesh.v1.EventEnvelope
}
var file_mesh_v1_bus_proto_depIdxs = []int32{
	4, // 0: mesh.v1.BusIngressRequest.envelope:type_name -> mesh.v1.EventEnvelope
	0, // 1: mesh.v1.BusIngressResponse.status:type_name -> mesh.v1.AckStatus
	2, // 2: mesh.v1.EventBus.Publish:output_type -> mesh.v1.BusIngressResponse
	4, // 3: mesh.v1.EventBus.Subscribe:output_type -> mesh.v1.EventEnvelope
	1, // 4: mesh.v1.EventBus.Publish:input_type -> mesh.v1.BusIngressRequest
	3, // 5: mesh.v1.EventBus.Subscribe:input_type -> mesh.v1.SubscriptionRequest
}

func init() { file_mesh_v1_bus_proto_init() }
func file_mesh_v1_bus_proto_init() {
	if File_mesh_v1_bus_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1_bus_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BusIngressRequest); i {
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
		file_mesh_v1_bus_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BusIngressResponse); i {
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
		file_mesh_v1_bus_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscriptionRequest); i {
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
			RawDescriptor: file_mesh_v1_bus_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mesh_v1_bus_proto_goTypes,
		DependencyIndexes: file_mesh_v1_bus_proto_depIdxs,
		EnumInfos:         file_mesh_v1_bus_proto_enumTypes,
		MessageInfos:      file_mesh_v1_bus_proto_msgTypes,
	}.Build()
	File_mesh_v1_bus_proto = out.File
	file_mesh_v1_bus_proto_rawDesc = nil
	file_mesh_v1_bus_proto_goTypes = nil
	file_mesh_v1_bus_proto_depIdxs = nil
}

var file_mesh_v1_bus_proto_rawDesc = file_mesh_v1_bus_proto_rawDescData

func boolPtr(v bool) *bool {
	return &v
}
