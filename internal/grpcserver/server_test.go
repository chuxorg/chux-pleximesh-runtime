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
	"google.golang.org/grpc/credentials/insecure"
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

	serverTLS, clientCreds := newTLSConfig(t)
	server, err := NewServer(Config{
		TLSConfig: serverTLS,
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

func TestServerRejectsInsecureClients(t *testing.T) {
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
	serverTLS, _ := newTLSConfig(t)

	server, err := NewServer(Config{
		TLSConfig: serverTLS,
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
	_, err = grpc.DialContext(ctx, server.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err == nil {
		t.Fatalf("expected TLS handshake failure")
	}
}

func newTLSConfig(t *testing.T) (*tls.Config, credentials.TransportCredentials) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	serverCert := tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
	}
	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS13,
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	clientCreds := credentials.NewClientTLSFromCert(pool, "localhost")
	return serverTLS, clientCreds
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
