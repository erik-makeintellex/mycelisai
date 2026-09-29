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

// ConnectAs establishes a named connection to the NATS server with automatic
// reconnects; an empty name means "Mycelis Core". MaxReconnects(-1) means
// unlimited, so transient infrastructure drops (k8s pod restart, bridge flap)
// heal without a Core restart. Distinct names separate chat-critical traffic
// from background observer/fanout lanes while keeping the same retry posture.
func ConnectAs(url, connectionName string) (*Client, error) {
	name := connectionName
	if name == "" {
		name = "Mycelis Core"
	}
	opts := []nats.Option{
		nats.Name(name),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // unlimited — heal automatically
		nats.PingInterval(20 * time.Second),
		nats.MaxPingsOutstanding(3),
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			subject := ""
			if sub != nil {
				subject = sub.Subject
			}
			if subject != "" {
				log.Printf("[nats] async error on %s: %v", subject, err)
				return
			}
			log.Printf("[nats] async error: %v", err)
		}),
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
