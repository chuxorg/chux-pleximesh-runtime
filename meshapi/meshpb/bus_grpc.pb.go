package meshpb

import (
	"context"
	"fmt"

	grpc "google.golang.org/grpc"
)

const _ = grpc.SupportPackageIsVersion7

type EventBusClient interface {
	Publish(ctx context.Context, opts ...grpc.CallOption) (EventBus_PublishClient, error)
	Subscribe(ctx context.Context, in *SubscriptionRequest, opts ...grpc.CallOption) (EventBus_SubscribeClient, error)
}

type eventBusClient struct {
	cc grpc.ClientConnInterface
}

func NewEventBusClient(cc grpc.ClientConnInterface) EventBusClient {
	return &eventBusClient{cc}
}

func (c *eventBusClient) Publish(ctx context.Context, opts ...grpc.CallOption) (EventBus_PublishClient, error) {
	stream, err := c.cc.NewStream(ctx, &EventBus_ServiceDesc.Streams[0], "/mesh.v1.EventBus/Publish", opts...)
	if err != nil {
		return nil, err
	}
	return &eventBusPublishClient{stream}, nil
}

type EventBus_PublishClient interface {
	Send(*BusIngressRequest) error
	Recv() (*BusIngressResponse, error)
	grpc.ClientStream
}

type eventBusPublishClient struct {
	grpc.ClientStream
}

func (x *eventBusPublishClient) Send(m *BusIngressRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *eventBusPublishClient) Recv() (*BusIngressResponse, error) {
	msg := new(BusIngressResponse)
	if err := x.ClientStream.RecvMsg(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *eventBusClient) Subscribe(ctx context.Context, in *SubscriptionRequest, opts ...grpc.CallOption) (EventBus_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &EventBus_ServiceDesc.Streams[1], "/mesh.v1.EventBus/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventBusSubscribeClient{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EventBus_SubscribeClient interface {
	Recv() (*EventEnvelope, error)
	grpc.ClientStream
}

type eventBusSubscribeClient struct {
	grpc.ClientStream
}

func (x *eventBusSubscribeClient) Recv() (*EventEnvelope, error) {
	msg := new(EventEnvelope)
	if err := x.ClientStream.RecvMsg(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

type EventBusServer interface {
	Publish(EventBus_PublishServer) error
	Subscribe(*SubscriptionRequest, EventBus_SubscribeServer) error
	mustEmbedUnimplementedEventBusServer()
}

type UnimplementedEventBusServer struct{}

func (UnimplementedEventBusServer) Publish(EventBus_PublishServer) error {
	return fmt.Errorf("method Publish not implemented")
}

func (UnimplementedEventBusServer) Subscribe(*SubscriptionRequest, EventBus_SubscribeServer) error {
	return fmt.Errorf("method Subscribe not implemented")
}

func (UnimplementedEventBusServer) mustEmbedUnimplementedEventBusServer() {}

type UnsafeEventBusServer interface {
	mustEmbedUnimplementedEventBusServer()
}

func RegisterEventBusServer(s grpc.ServiceRegistrar, srv EventBusServer) {
	s.RegisterService(&EventBus_ServiceDesc, srv)
}

func _EventBus_Publish_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EventBusServer).Publish(&eventBusPublishServer{stream})
}

type EventBus_PublishServer interface {
	Send(*BusIngressResponse) error
	Recv() (*BusIngressRequest, error)
	grpc.ServerStream
}

type eventBusPublishServer struct {
	grpc.ServerStream
}

func (x *eventBusPublishServer) Send(m *BusIngressResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *eventBusPublishServer) Recv() (*BusIngressRequest, error) {
	msg := new(BusIngressRequest)
	if err := x.ServerStream.RecvMsg(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func _EventBus_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscriptionRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EventBusServer).Subscribe(m, &eventBusSubscribeServer{stream})
}

type EventBus_SubscribeServer interface {
	Send(*EventEnvelope) error
	grpc.ServerStream
}

type eventBusSubscribeServer struct {
	grpc.ServerStream
}

func (x *eventBusSubscribeServer) Send(m *EventEnvelope) error {
	return x.ServerStream.SendMsg(m)
}

// EventBus_ServiceDesc describes the EventBus service for registration.
var EventBus_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mesh.v1.EventBus",
	HandlerType: (*EventBusServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Publish",
			Handler:       _EventBus_Publish_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Subscribe",
			Handler:       _EventBus_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "mesh/v1/bus.proto",
}
