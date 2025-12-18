module github.com/chuxorg/chux-agent-mesh

go 1.24.0

require (
	google.golang.org/grpc v0.0.0
	google.golang.org/protobuf v1.34.1
)

replace google.golang.org/grpc => ./third_party/grpcshim
