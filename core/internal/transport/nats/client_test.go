package nats

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

var transportTestNATSPort int32 = 14620 + int32(time.Now().UnixNano()%1000)

func TestConnectAs_UsesDistinctConnectionNamesForSplitLanes(t *testing.T) {
	opts := &natsserver.Options{Host: "127.0.0.1", Port: nextTransportTestNATSPort(t)}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	t.Cleanup(func() {
		srv.Shutdown()
		srv.WaitForShutdown()
	})

	coreClient, err := Connect(srv.ClientURL())
	if err != nil {
		t.Fatalf("connect core client: %v", err)
	}
	t.Cleanup(func() {
		_ = coreClient.Drain()
	})

	observerClient, err := ConnectAs(srv.ClientURL(), "Mycelis Observer")
	if err != nil {
		t.Fatalf("connect observer client: %v", err)
	}
	t.Cleanup(func() {
		_ = observerClient.Drain()
	})

	if got := coreClient.Conn.Opts.Name; got != "Mycelis Core" {
		t.Fatalf("core connection name = %q, want %q", got, "Mycelis Core")
	}
	if got := observerClient.Conn.Opts.Name; got != "Mycelis Observer" {
		t.Fatalf("observer connection name = %q, want %q", got, "Mycelis Observer")
	}
	if coreClient.Conn == observerClient.Conn {
		t.Fatal("expected separate NATS connections for core and observer lanes")
	}
	if !coreClient.Conn.IsConnected() || !observerClient.Conn.IsConnected() {
		t.Fatal("expected both split-lane connections to be live")
	}
}

func nextTransportTestNATSPort(t *testing.T) int {
	t.Helper()
	for range 200 {
		port := int(atomic.AddInt32(&transportTestNATSPort, 1))
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			_ = ln.Close()
			return port
		}
	}
	t.Fatal("no available low NATS test port")
	return 0
}
