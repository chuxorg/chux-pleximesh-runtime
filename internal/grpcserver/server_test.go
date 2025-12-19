package grpcserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/ems"
	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
	"github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"
	"github.com/chuxorg/chux-agent-mesh/pkg/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestServerPublishesToMultipleSubscribers(t *testing.T) {
	router := messagebus.NewRouter()
	now := time.Unix(1_700_000_300, 0).UTC()

	clock := func() time.Time { return now }
	km := ems.NewKeyManager(ems.WithTimeSource(clock))
	material := fixedMaterial(0x33)
	rootKey, err := km.RegisterKey("publisher-1", material[:])
	if err != nil {
		t.Fatalf("register key: %v", err)
	}
	scheduler := ems.NewSliceScheduler(ems.WithSchedulerClock(clock))
	verifier := ems.NewVerifier(km, scheduler, ems.WithVerifierClock(clock))

	serverTLS, clientCreds, _, clientCAPool := newMutualTLSConfigs(t)
	server, err := NewServer(Config{
		TLSConfig: serverTLS,
		ClientCAs: clientCAPool,
		Router:    router,
		Verifier:  verifier,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Stop(stopCtx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, server.Addr(), grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := meshpb.NewEventBusClient(conn)
	type subscriber struct {
		stream   meshpb.EventBus_SubscribeClient
		received chan *meshpb.EventEnvelope
	}
	var subs []*subscriber
	for i := 0; i < 3; i++ {
		stream, err := client.Subscribe(ctx, &meshpb.SubscriptionRequest{AgentId: fmt.Sprintf("sub-%d", i)})
		if err != nil {
			t.Fatalf("subscribe %d: %v", i, err)
		}
		s := &subscriber{
			stream:   stream,
			received: make(chan *meshpb.EventEnvelope, 10),
		}
		subs = append(subs, s)
		go func(sub *subscriber) {
			for {
				env, err := sub.stream.Recv()
				if err != nil {
					close(sub.received)
					return
				}
				sub.received <- env
			}
		}(s)
	}

	pubStream, err := client.Publish(ctx)
	if err != nil {
		t.Fatalf("publish stream: %v", err)
	}

	events := []event.Envelope{
		newEvent("evt-1", now),
		newEvent("evt-2", now.Add(1*time.Second)),
		newEvent("evt-3", now.Add(2*time.Second)),
	}
	slice := scheduler.SliceAt(now)
	for _, env := range events {
		env.SourceAgent.AgentID = "publisher-1"
		env.SourceAgent.Role = "engineer"
		env.Runtime.RuntimeVersion = "mesh"
		env.Runtime.InitKitVersion = "initkit"
		env.Signature = ""
		signEnvelope(t, &env, rootKey.Material, slice)
		req := &meshpb.BusIngressRequest{Envelope: envelopeToProto(env)}
		if err := pubStream.Send(req); err != nil {
			t.Fatalf("send event: %v", err)
		}
		resp, err := pubStream.Recv()
		if err != nil {
			t.Fatalf("recv ack: %v", err)
		}
		if resp.GetStatus() != meshpb.AckStatus_ACK_STATUS_ACCEPTED {
			t.Fatalf("expected accepted ack, got %v", resp.GetStatus())
		}
	}
	_ = pubStream.CloseSend()

	for i, sub := range subs {
		for _, expected := range events {
			select {
			case env := <-sub.received:
				if env.GetEventId() != expected.EventID {
					t.Fatalf("subscriber %d received %s, expected %s", i, env.GetEventId(), expected.EventID)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("subscriber %d timed out", i)
			}
		}
	}
}

func TestServerRejectsClientsWithoutCertificates(t *testing.T) {
	router := messagebus.NewRouter()
	now := time.Now().UTC()
	km := ems.NewKeyManager(ems.WithTimeSource(func() time.Time { return now }))
	material := fixedMaterial(0x44)
	_, err := km.RegisterKey("publisher-1", material[:])
	if err != nil {
		t.Fatalf("register key: %v", err)
	}
	scheduler := ems.NewSliceScheduler()
	verifier := ems.NewVerifier(km, scheduler)
	serverTLS, _, noClientCreds, clientCAPool := newMutualTLSConfigs(t)

	server, err := NewServer(Config{
		TLSConfig: serverTLS,
		ClientCAs: clientCAPool,
		Router:    router,
		Verifier:  verifier,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Stop(stopCtx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = grpc.DialContext(ctx, server.Addr(), grpc.WithTransportCredentials(noClientCreds), grpc.WithBlock())
	if err == nil {
		t.Fatalf("expected mutual TLS handshake failure for missing client certificate")
	}
}

func newMutualTLSConfigs(t *testing.T) (*tls.Config, credentials.TransportCredentials, credentials.TransportCredentials, *x509.CertPool) {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create server cert: %v", err)
	}
	serverCert := tls.Certificate{
		Certificate: [][]byte{serverDER, caDER},
		PrivateKey:  serverKey,
	}
	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate client key: %v", err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create client cert: %v", err)
	}
	clientCert := tls.Certificate{
		Certificate: [][]byte{clientDER, caDER},
		PrivateKey:  clientKey,
	}
	authClientTLS := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
	}
	noClientTLS := &tls.Config{
		RootCAs:    caPool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS13,
	}
	return serverTLS, credentials.NewTLS(authClientTLS), credentials.NewTLS(noClientTLS), caPool
}

const (
	testKeySize           = 32
	testSignaturePrefix   = "ems-hmac-v1:"
	contextFieldSeparator = "|"
)

func signEnvelope(t *testing.T, env *event.Envelope, material [testKeySize]byte, slice ems.SliceIndex) {
	t.Helper()
	env.Signature = ""
	canonical, err := canonicalizeForTest(*env)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	ctx := deriveContextForTest(*env, slice)
	ephemeral := hmac.New(sha256.New, material[:])
	ephemeral.Write(ctx)
	ephemeralKey := ephemeral.Sum(nil)
	mac := hmac.New(sha256.New, ephemeralKey)
	mac.Write(canonical)
	sig := mac.Sum(nil)
	env.Signature = testSignaturePrefix + base64.StdEncoding.EncodeToString(sig)
}

func newEvent(id string, ts time.Time) event.Envelope {
	return event.Envelope{
		EventID:       id,
		Type:          event.TypeTaskCreated,
		Domain:        event.DomainTask,
		Timestamp:     ts,
		CorrelationID: "corr-" + id,
		Payload:       []byte(`{"task":"` + id + `"}`),
		Runtime: event.RuntimeContext{
			RuntimeVersion: "mesh",
			InitKitVersion: "initkit",
		},
		SourceAgent: event.AgentDescriptor{
			AgentID: "publisher-1",
			Role:    "engineer",
		},
	}
}

func fixedMaterial(v byte) [testKeySize]byte {
	var material [testKeySize]byte
	for i := range material {
		material[i] = v
	}
	return material
}

func canonicalizeForTest(env event.Envelope) ([]byte, error) {
	copyEnv := env
	copyEnv.Signature = ""
	copyEnv.Timestamp = copyEnv.Timestamp.UTC()
	if len(env.Payload) > 0 {
		copyEnv.Payload = append([]byte(nil), env.Payload...)
	}
	return json.Marshal(copyEnv)
}

func deriveContextForTest(env event.Envelope, slice ems.SliceIndex) []byte {
	builder := strings.Builder{}
	builder.WriteString(env.SourceAgent.AgentID)
	builder.WriteString(contextFieldSeparator)
	builder.WriteString(string(env.Domain))
	builder.WriteString(contextFieldSeparator)
	builder.WriteString(string(env.Type))
	builder.WriteString(contextFieldSeparator)
	builder.WriteString(env.CorrelationID)
	builder.WriteString(contextFieldSeparator)
	builder.WriteString(strconv.FormatInt(int64(slice), 10))
	return []byte(builder.String())
}
