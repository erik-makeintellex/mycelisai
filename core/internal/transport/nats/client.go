package nats

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Client wraps the standard NATS connection
type Client struct {
	Conn *nats.Conn
}

// Connect establishes a connection to the NATS server with automatic reconnects.
// MaxReconnects(-1) means unlimited — the process will keep retrying indefinitely
// so that transient infrastructure drops (k8s pod restart, bridge flap) are healed
// without requiring a Core restart.
func Connect(url string) (*Client, error) {
	opts := []nats.Option{
		nats.Name("Mycelis Core"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // unlimited — heal automatically
		nats.PingInterval(20 * time.Second),
		nats.MaxPingsOutstanding(3),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("[nats] disconnected: %v — will retry", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[nats] reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("[nats] connection permanently closed")
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{Conn: nc}, nil
}

// Drain cleans up the connection gracefully
func (c *Client) Drain() error {
	return c.Conn.Drain()
}
