package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// SupportPackageIsVersion7 is used by generated stubs to assert compatibility.
const SupportPackageIsVersion7 = 1

// CallOption configures call behaviors.
type CallOption interface{}

// ClientConnInterface represents a gRPC client connection.
type ClientConnInterface interface {
	Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...CallOption) error
	NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (ClientStream, error)
}

// ClientStream defines the minimum client streaming behavior needed by generated stubs.
type ClientStream interface {
	Header() (metadata.MD, error)
	Trailer() metadata.MD
	CloseSend() error
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// ServerStream defines the minimum server streaming behavior needed by generated stubs.
type ServerStream interface {
	SetHeader(metadata.MD) error
	SendHeader(metadata.MD) error
	SetTrailer(metadata.MD)
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// StreamHandler handles server-side streaming RPCs.
type StreamHandler func(srv interface{}, stream ServerStream) error

// UnaryHandler handles unary RPC requests.
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// MethodDesc describes a unary RPC method.
type MethodDesc struct {
	MethodName string
	Handler    UnaryHandler
}

// StreamDesc describes a streaming RPC method.
type StreamDesc struct {
	StreamName    string
	Handler       StreamHandler
	ServerStreams bool
	ClientStreams bool
}

// ServiceRegistrar registers service implementations.
type ServiceRegistrar interface {
	RegisterService(*ServiceDesc, interface{})
}

// ServiceDesc describes a service for registrar implementations.
type ServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Methods     []MethodDesc
	Streams     []StreamDesc
	Metadata    interface{}
}
