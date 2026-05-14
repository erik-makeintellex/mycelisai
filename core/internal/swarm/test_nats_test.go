package swarm

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var swarmTestServerPort int32 = 12000 + int32(time.Now().UnixNano()%1000)

func startTestNATS(t *testing.T) (*natsserver.Server, *nats.Conn) {
	t.Helper()

	opts := &natsserver.Options{
		Host: "127.0.0.1",
		Port: reserveSwarmTestPort(t),
	}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}

	nc, err := nats.Connect(srv.ClientURL(), nats.Dialer(&net.Dialer{Timeout: 2 * time.Second}))
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

func reserveSwarmTestPort(t *testing.T) int {
	t.Helper()
	for range 12000 {
		port := int(atomic.AddInt32(&swarmTestServerPort, 1))
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			_ = ln.Close()
			return port
		}
	}
	t.Fatal("no available low server test port")
	return 0
}
