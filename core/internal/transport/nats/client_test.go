package nats

import (
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

func TestConnectAs_UsesDistinctConnectionNamesForSplitLanes(t *testing.T) {
	opts := &natsserver.Options{Host: "127.0.0.1", Port: -1}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	t.Cleanup(srv.Shutdown)

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
