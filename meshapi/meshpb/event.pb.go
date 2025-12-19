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

type AgentDescriptor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AgentId string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	Role    string `protobuf:"bytes,2,opt,name=role,proto3" json:"role,omitempty"`
}

func (x *AgentDescriptor) Reset() {
	*x = AgentDescriptor{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AgentDescriptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgentDescriptor) ProtoMessage() {}

func (x *AgentDescriptor) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*AgentDescriptor) Descriptor() ([]byte, []int) {
	return file_mesh_v1_event_proto_rawDescGZIP(), []int{0}
}

func (x *AgentDescriptor) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

func (x *AgentDescriptor) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

type TargetDescriptor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AgentId string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	Scope   string `protobuf:"bytes,2,opt,name=scope,proto3" json:"scope,omitempty"`
}

func (x *TargetDescriptor) Reset() {
	*x = TargetDescriptor{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TargetDescriptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TargetDescriptor) ProtoMessage() {}

func (x *TargetDescriptor) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TargetDescriptor) Descriptor() ([]byte, []int) {
	return file_mesh_v1_event_proto_rawDescGZIP(), []int{1}
}

func (x *TargetDescriptor) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

func (x *TargetDescriptor) GetScope() string {
	if x != nil {
		return x.Scope
	}
	return ""
}

type RuntimeContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RuntimeVersion string `protobuf:"bytes,1,opt,name=runtime_version,json=runtimeVersion,proto3" json:"runtime_version,omitempty"`
	InitkitVersion string `protobuf:"bytes,2,opt,name=initkit_version,json=initkitVersion,proto3" json:"initkit_version,omitempty"`
}

func (x *RuntimeContext) Reset() {
	*x = RuntimeContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_event_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RuntimeContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RuntimeContext) ProtoMessage() {}

func (x *RuntimeContext) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_event_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RuntimeContext) Descriptor() ([]byte, []int) {
	return file_mesh_v1_event_proto_rawDescGZIP(), []int{2}
}

func (x *RuntimeContext) GetRuntimeVersion() string {
	if x != nil {
		return x.RuntimeVersion
	}
	return ""
}

func (x *RuntimeContext) GetInitkitVersion() string {
	if x != nil {
		return x.InitkitVersion
	}
	return ""
}

type EventEnvelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId       string            `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	EventType     string            `protobuf:"bytes,2,opt,name=event_type,json=eventType,proto3" json:"event_type,omitempty"`
	EventDomain   string            `protobuf:"bytes,3,opt,name=event_domain,json=eventDomain,proto3" json:"event_domain,omitempty"`
	SourceAgent   *AgentDescriptor  `protobuf:"bytes,4,opt,name=source_agent,json=sourceAgent,proto3" json:"source_agent,omitempty"`
	Target        *TargetDescriptor `protobuf:"bytes,5,opt,name=target,proto3" json:"target,omitempty"`
	Timestamp     string            `protobuf:"bytes,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	CorrelationId string            `protobuf:"bytes,7,opt,name=correlation_id,json=correlationId,proto3" json:"correlation_id,omitempty"`
	Payload       []byte            `protobuf:"bytes,8,opt,name=payload,proto3" json:"payload,omitempty"`
	Signature     string            `protobuf:"bytes,9,opt,name=signature,proto3" json:"signature,omitempty"`
	Runtime       *RuntimeContext   `protobuf:"bytes,10,opt,name=runtime,proto3" json:"runtime,omitempty"`
}

func (x *EventEnvelope) Reset() {
	*x = EventEnvelope{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1_event_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventEnvelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventEnvelope) ProtoMessage() {}

func (x *EventEnvelope) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1_event_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*EventEnvelope) Descriptor() ([]byte, []int) {
	return file_mesh_v1_event_proto_rawDescGZIP(), []int{3}
}

func (x *EventEnvelope) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *EventEnvelope) GetEventType() string {
	if x != nil {
		return x.EventType
	}
	return ""
}

func (x *EventEnvelope) GetEventDomain() string {
	if x != nil {
		return x.EventDomain
	}
	return ""
}

func (x *EventEnvelope) GetSourceAgent() *AgentDescriptor {
	if x != nil {
		return x.SourceAgent
	}
	return nil
}

func (x *EventEnvelope) GetTarget() *TargetDescriptor {
	if x != nil {
		return x.Target
	}
	return nil
}

func (x *EventEnvelope) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

func (x *EventEnvelope) GetCorrelationId() string {
	if x != nil {
		return x.CorrelationId
	}
	return ""
}

func (x *EventEnvelope) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *EventEnvelope) GetSignature() string {
	if x != nil {
		return x.Signature
	}
	return ""
}

func (x *EventEnvelope) GetRuntime() *RuntimeContext {
	if x != nil {
		return x.Runtime
	}
	return nil
}

var File_mesh_v1_event_proto protoreflect.FileDescriptor

var file_mesh_v1_event_proto_rawDescOnce sync.Once
var file_mesh_v1_event_proto_rawDescData = buildMeshV1EventProtoRawDesc()

func buildMeshV1EventProtoRawDesc() []byte {
	fd := &descriptorpb.FileDescriptorProto{
		Syntax:  strPtr("proto3"),
		Name:    strPtr("mesh/v1/event.proto"),
		Package: strPtr("mesh.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: strPtr("github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: strPtr("AgentDescriptor"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("agent_id"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("agentId"),
					},
					{
						Name:     strPtr("role"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("role"),
					},
				},
			},
			{
				Name: strPtr("TargetDescriptor"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("agent_id"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("agentId"),
					},
					{
						Name:     strPtr("scope"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("scope"),
					},
				},
			},
			{
				Name: strPtr("RuntimeContext"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("runtime_version"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("runtimeVersion"),
					},
					{
						Name:     strPtr("initkit_version"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("initkitVersion"),
					},
				},
			},
			{
				Name: strPtr("EventEnvelope"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     strPtr("event_id"),
						Number:   int32Ptr(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("eventId"),
					},
					{
						Name:     strPtr("event_type"),
						Number:   int32Ptr(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("eventType"),
					},
					{
						Name:     strPtr("event_domain"),
						Number:   int32Ptr(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("eventDomain"),
					},
					{
						Name:     strPtr("source_agent"),
						Number:   int32Ptr(4),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: strPtr(".mesh.v1.AgentDescriptor"),
						JsonName: strPtr("sourceAgent"),
					},
					{
						Name:     strPtr("target"),
						Number:   int32Ptr(5),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: strPtr(".mesh.v1.TargetDescriptor"),
						JsonName: strPtr("target"),
					},
					{
						Name:     strPtr("timestamp"),
						Number:   int32Ptr(6),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("timestamp"),
					},
					{
						Name:     strPtr("correlation_id"),
						Number:   int32Ptr(7),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("correlationId"),
					},
					{
						Name:     strPtr("payload"),
						Number:   int32Ptr(8),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum(),
						JsonName: strPtr("payload"),
					},
					{
						Name:     strPtr("signature"),
						Number:   int32Ptr(9),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: strPtr("signature"),
					},
					{
						Name:     strPtr("runtime"),
						Number:   int32Ptr(10),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: strPtr(".mesh.v1.RuntimeContext"),
						JsonName: strPtr("runtime"),
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

func file_mesh_v1_event_proto_rawDescGZIP() []byte {
	file_mesh_v1_event_proto_rawDescOnce.Do(func() {
		file_mesh_v1_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1_event_proto_rawDescData)
	})
	return file_mesh_v1_event_proto_rawDescData
}

var file_mesh_v1_event_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_mesh_v1_event_proto_goTypes = []interface{}{
	(*AgentDescriptor)(nil),  // 0: mesh.v1.AgentDescriptor
	(*TargetDescriptor)(nil), // 1: mesh.v1.TargetDescriptor
	(*RuntimeContext)(nil),   // 2: mesh.v1.RuntimeContext
	(*EventEnvelope)(nil),    // 3: mesh.v1.EventEnvelope
}
var file_mesh_v1_event_proto_depIdxs = []int32{
	0, // 0: mesh.v1.EventEnvelope.source_agent:type_name -> mesh.v1.AgentDescriptor
	1, // 1: mesh.v1.EventEnvelope.target:type_name -> mesh.v1.TargetDescriptor
	2, // 2: mesh.v1.EventEnvelope.runtime:type_name -> mesh.v1.RuntimeContext
}

func init() { file_mesh_v1_event_proto_init() }
func file_mesh_v1_event_proto_init() {
	if File_mesh_v1_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AgentDescriptor); i {
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
		file_mesh_v1_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TargetDescriptor); i {
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
		file_mesh_v1_event_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RuntimeContext); i {
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
		file_mesh_v1_event_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventEnvelope); i {
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
			RawDescriptor: file_mesh_v1_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1_event_proto_goTypes,
		DependencyIndexes: file_mesh_v1_event_proto_depIdxs,
		MessageInfos:      file_mesh_v1_event_proto_msgTypes,
	}.Build()
	File_mesh_v1_event_proto = out.File
	file_mesh_v1_event_proto_rawDesc = nil
	file_mesh_v1_event_proto_goTypes = nil
	file_mesh_v1_event_proto_depIdxs = nil
}

var file_mesh_v1_event_proto_rawDesc = file_mesh_v1_event_proto_rawDescData

func strPtr(s string) *string {
	return &s
}

func int32Ptr(v int32) *int32 {
	return &v
}
