package overseer

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var overseerTestNATSPort int32 = 14420 + int32(time.Now().UnixNano()%1000)
var overseerTestNATSClientPort int32 = 35000 + int32(time.Now().UnixNano()%5000)

func startTestNATS(t *testing.T) (*natsserver.Server, *nats.Conn) {
	t.Helper()
	opts := &natsserver.Options{Host: "127.0.0.1", Port: nextOverseerTestNATSPort(t)}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL(), nats.Dialer(&net.Dialer{
		LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: nextOverseerTestNATSClientPort(t)},
		Timeout:   2 * time.Second,
	}))
	if err != nil {
		srv.Shutdown()
		srv.WaitForShutdown()
		t.Fatalf("nats connect: %v", err)
	}
	t.Cleanup(func() {
		nc.Close()
		srv.Shutdown()
		srv.WaitForShutdown()
	})
	return srv, nc
}

func nextOverseerTestNATSPort(t *testing.T) int {
	return nextOverseerTestTCPPort(t, &overseerTestNATSPort)
}

func nextOverseerTestNATSClientPort(t *testing.T) int {
	return nextOverseerTestTCPPort(t, &overseerTestNATSClientPort)
}

func nextOverseerTestTCPPort(t *testing.T, counter *int32) int {
	t.Helper()
	for range 200 {
		port := int(atomic.AddInt32(counter, 1))
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			_ = ln.Close()
			return port
		}
	}
	t.Fatal("no available low NATS test port")
	return 0
}
