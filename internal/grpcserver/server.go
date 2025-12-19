package grpcserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"

	"github.com/chuxorg/chux-agent-mesh/internal/ems"
	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
	"github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config defines how the gRPC server should be bootstrapped.
type Config struct {
	Address   string
	Listener  net.Listener
	TLSConfig *tls.Config
	Router    *messagebus.Router
	Verifier  *ems.Verifier
}

// Server wraps a gRPC EventBus server instance.
type Server struct {
	cfg      Config
	grpc     *grpc.Server
	listener net.Listener
	serveErr chan error
	started  uint32
}

// NewServer constructs a TLS-enabled gRPC server using the provided configuration.
func NewServer(cfg Config) (*Server, error) {
	if cfg.Router == nil {
		return nil, errors.New("grpcserver: router is required")
	}
	if cfg.Verifier == nil {
		return nil, errors.New("grpcserver: verifier is required")
	}
	if cfg.TLSConfig == nil {
		return nil, errors.New("grpcserver: TLS config is required")
	}
	tlsConfig := cfg.TLSConfig.Clone()
	if tlsConfig == nil {
		tlsConfig = cfg.TLSConfig
	}
	if len(tlsConfig.Certificates) == 0 && tlsConfig.GetCertificate == nil {
		return nil, errors.New("grpcserver: TLS certificate is required")
	}
	if tlsConfig.MinVersion == 0 {
		tlsConfig.MinVersion = tls.VersionTLS13
	}
	creds := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	meshpb.RegisterEventBusServer(grpcServer, &busService{
		router:   cfg.Router,
		verifier: cfg.Verifier,
	})
	return &Server{
		cfg:      cfg,
		grpc:     grpcServer,
		serveErr: make(chan error, 1),
	}, nil
}

// Start begins serving the EventBus over TLS.
func (s *Server) Start() error {
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New("grpcserver: server already started")
	}
	var err error
	s.listener = s.cfg.Listener
	if s.listener == nil {
		addr := s.cfg.Address
		if addr == "" {
			addr = "127.0.0.1:0"
		}
		s.listener, err = net.Listen("tcp", addr)
		if err != nil {
			atomic.StoreUint32(&s.started, 0)
			return err
		}
	}
	go func() {
		err := s.grpc.Serve(s.listener)
		s.serveErr <- err
		close(s.serveErr)
	}()
	return nil
}

// Addr returns the listener address if the server has started.
func (s *Server) Addr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// Stop gracefully stops the server, falling back to a hard stop when the context expires.
func (s *Server) Stop(ctx context.Context) error {
	if atomic.LoadUint32(&s.started) == 0 {
		return nil
	}
	done := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(done)
	}()
	var stopErr error
	select {
	case <-done:
	case <-ctx.Done():
		stopErr = ctx.Err()
		s.grpc.Stop()
		<-done
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	atomic.StoreUint32(&s.started, 0)
	if err := <-s.serveErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return stopErr
}
