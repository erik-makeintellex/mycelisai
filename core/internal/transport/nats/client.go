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

// Connect establishes a connection to the NATS server with automatic reconnects
func Connect(url string) (*Client, error) {
	opts := []nats.Option{
		nats.Name("Mycelis Core"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(50),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("Disconnected from NATS: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to NATS [%s]", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
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
