package reactive

import (
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var testNATSPort int32 = 14320

func startTestNATS(t *testing.T) (*natsserver.Server, *nats.Conn) {
	t.Helper()
	opts := &natsserver.Options{Host: "127.0.0.1", Port: int(atomic.AddInt32(&testNATSPort, 1))}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(
		srv.ClientURL(),
		nats.Name("reactive-engine-test"),
		nats.Timeout(5*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
		nats.ReconnectWait(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("nats connect: %v", err)
	}
	return srv, nc
}
