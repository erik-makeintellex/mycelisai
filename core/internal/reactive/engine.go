// Package reactive implements the mission profile reactive subscription engine.
// Profiles with subscriptions watch specific NATS topic patterns and route
// received messages to Soma for evaluation — Soma decides whether to engage
// (reactive-watch, not forced auto-chain).
package reactive

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

// ProfileSubscription defines a NATS topic pattern a profile watches.
type ProfileSubscription struct {
	Topic     string `json:"topic"`              // e.g. "swarm.team.research-team.*"
	Condition string `json:"condition,omitempty"` // optional filter (reserved for future use)
}

// ReactHandler is called when a subscribed NATS message arrives.
// profileID identifies which profile matched. topic is the concrete subject.
// msg is the raw payload. The handler decides whether and how to respond.
type ReactHandler func(profileID, topic string, msg []byte)

// Engine manages reactive NATS subscriptions for all active mission profiles.
type Engine struct {
	nc      *nats.Conn
	db      *sql.DB                          // optional — used for ReactivateFromDB
	subs    map[string][]*nats.Subscription  // profileID → subscriptions
	handler ReactHandler
	mu      sync.Mutex
}

// New creates a reactive Engine. nc may be nil (graceful degradation).
// handler is called on each matching NATS message.
func New(nc *nats.Conn, handler ReactHandler) *Engine {
	e := &Engine{
		nc:      nc,
		subs:    make(map[string][]*nats.Subscription),
		handler: handler,
	}
	// Register reconnect handler so subscriptions survive NATS drops.
	// The nats.go library auto-restores subscriptions made before the disconnect;
	// ReactivateFromDB handles any that were only stored in the DB (not live yet).
	if nc != nil {
		nc.SetReconnectHandler(func(c *nats.Conn) {
			log.Printf("[reactive] NATS reconnected — refreshing active profile subscriptions")
			if e.db != nil {
				if err := e.ReactivateFromDB(context.Background()); err != nil {
					log.Printf("[reactive] ReactivateFromDB after reconnect: %v", err)
				}
			}
		})
	}
	return e
}

// SetDB wires a database handle so the engine can re-subscribe active profiles
// after a NATS reconnect.
func (e *Engine) SetDB(db *sql.DB) {
	e.mu.Lock()
	e.db = db
	e.mu.Unlock()
}

// ReactivateFromDB loads all active mission profiles from the database and
// re-establishes their NATS subscriptions. Safe to call multiple times.
func (e *Engine) ReactivateFromDB(ctx context.Context) error {
	if e.db == nil || e.nc == nil {
		return nil
	}

	rows, err := e.db.QueryContext(ctx,
		`SELECT id, subscriptions FROM mission_profiles
		 WHERE is_active = true AND tenant_id = 'default'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var profileID string
		var subsJSON []byte
		if err := rows.Scan(&profileID, &subsJSON); err != nil {
			log.Printf("[reactive] scan error: %v", err)
			continue
		}
		var subs []ProfileSubscription
		if err := json.Unmarshal(subsJSON, &subs); err != nil || len(subs) == 0 {
			continue
		}
		if subErr := e.Subscribe(profileID, subs); subErr != nil {
			log.Printf("[reactive] re-subscribe profile %s: %v", profileID, subErr)
		}
	}
	return rows.Err()
}

// Subscribe registers NATS subscriptions for a profile.
// Existing subscriptions for the profileID are replaced.
func (e *Engine) Subscribe(profileID string, topics []ProfileSubscription) error {
	if e.nc == nil {
		log.Printf("[reactive] NATS unavailable — profile %s subscriptions deferred", profileID)
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Drain existing subscriptions for this profile
	e.drainLocked(profileID)

	var newSubs []*nats.Subscription
	for _, t := range topics {
		if t.Topic == "" {
			continue
		}
		topic := t.Topic // capture for closure
		sub, err := e.nc.Subscribe(topic, func(msg *nats.Msg) {
			if e.handler != nil {
				e.handler(profileID, msg.Subject, msg.Data)
			}
		})
		if err != nil {
			log.Printf("[reactive] failed to subscribe profile %s to %s: %v", profileID, topic, err)
			continue
		}
		newSubs = append(newSubs, sub)
		log.Printf("[reactive] profile %s subscribed to %s", profileID, topic)
	}

	if len(newSubs) > 0 {
		e.subs[profileID] = newSubs
	}
	return nil
}

// Unsubscribe drains all NATS subscriptions for a profile.
func (e *Engine) Unsubscribe(profileID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.drainLocked(profileID)
}

// drainLocked drains subscriptions for profileID. Must hold e.mu.
func (e *Engine) drainLocked(profileID string) {
	for _, sub := range e.subs[profileID] {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("[reactive] unsubscribe error for profile %s: %v", profileID, err)
		}
	}
	delete(e.subs, profileID)
}

// Connected reports whether the underlying NATS connection is live.
func (e *Engine) Connected() bool {
	if e == nil || e.nc == nil {
		return false
	}
	return e.nc.IsConnected()
}

// ActiveSubscriptionCount returns the total number of live NATS subscriptions.
func (e *Engine) ActiveSubscriptionCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := 0
	for _, subs := range e.subs {
		n += len(subs)
	}
	return n
}

// Close unsubscribes all profiles. Called on server shutdown.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for profileID := range e.subs {
		e.drainLocked(profileID)
	}
}
