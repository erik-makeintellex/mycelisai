package cognitive

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EmbeddingDimensions is the context_vectors.embedding width. A provider
// that returns any other width cannot serve semantic recall.
const EmbeddingDimensions = 768

// EmbeddingNegativeCacheTTL bounds how long a failed probe or embed call
// short-circuits further Embed calls, so a dead engine costs no round trips.
const EmbeddingNegativeCacheTTL = 5 * time.Minute

// embeddingPositiveCacheTTL keeps a successful probe from re-running every turn.
const embeddingPositiveCacheTTL = 5 * time.Minute

const embeddingProbeText = "embedding availability probe"

// ErrEmbeddingUnavailable is returned without calling any provider while the
// negative cache is active.
var ErrEmbeddingUnavailable = errors.New("embedding engine unavailable")

// EmbeddingWidthMismatch is the status code for an engine whose vectors do
// not fit the store.
const EmbeddingWidthMismatch = "embedding_width_mismatch"

// EmbeddingStatus is the exported availability snapshot for status surfaces.
type EmbeddingStatus struct {
	Available bool      `json:"available"`
	Code      string    `json:"code,omitempty"` // embedding_unavailable | embedding_width_mismatch
	Reason    string    `json:"reason,omitempty"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
}

type embedAvailability struct {
	mu        sync.Mutex
	known     bool
	available bool
	code      string
	reason    string
	checkedAt time.Time
	now       func() time.Time
}

func (e *embedAvailability) clock() time.Time {
	if e.now != nil {
		return e.now()
	}
	return time.Now()
}

// blocked reports whether a recent failure should short-circuit Embed.
func (e *embedAvailability) blocked() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.known && !e.available && e.clock().Sub(e.checkedAt) < EmbeddingNegativeCacheTTL
}

func (e *embedAvailability) record(available bool, reason string) {
	code := ""
	if !available {
		code = "embedding_unavailable"
	}
	e.recordCode(available, code, reason)
}

func (e *embedAvailability) recordCode(available bool, code, reason string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.known, e.available, e.code, e.reason, e.checkedAt = true, available, code, reason, e.clock()
}

// fresh returns the cached verdict while it is inside its TTL.
func (e *embedAvailability) fresh() (bool, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.known {
		return false, false
	}
	ttl := embeddingPositiveCacheTTL
	if !e.available {
		ttl = EmbeddingNegativeCacheTTL
	}
	if e.clock().Sub(e.checkedAt) >= ttl {
		return false, false
	}
	return e.available, true
}

// EmbeddingAvailable probes the embedding engine at most once per TTL. A
// provider that errors or returns a vector of the wrong width is unavailable.
func (r *Router) EmbeddingAvailable(ctx context.Context) bool {
	if r == nil || r.Config == nil {
		return false
	}
	if available, ok := r.embedState.fresh(); ok {
		return available
	}
	vec, err := r.Embed(ctx, embeddingProbeText, "")
	return err == nil && len(vec) == EmbeddingDimensions
}

// recordWidth records a successful call; a vector of the wrong width cannot
// be stored, so it counts as unavailable.
func (e *embedAvailability) recordWidth(width int) {
	if width != EmbeddingDimensions {
		e.recordCode(false, EmbeddingWidthMismatch, fmt.Sprintf("embedding width %d does not match the %d-dim store", width, EmbeddingDimensions))
		return
	}
	e.record(true, "")
}

// EmbeddingStatus returns the last probe or embed verdict without probing.
func (r *Router) EmbeddingStatus() EmbeddingStatus {
	if r == nil {
		return EmbeddingStatus{Reason: "cognitive engine offline"}
	}
	r.embedState.mu.Lock()
	defer r.embedState.mu.Unlock()
	return EmbeddingStatus{Available: r.embedState.known && r.embedState.available, Code: r.embedState.code,
		Reason: r.embedState.reason, CheckedAt: r.embedState.checkedAt}
}

// embedProvider resolves the embed adapter: the "embed" profile first, then
// any adapter that implements EmbedProvider.
func (r *Router) embedProvider() (EmbedProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if providerID, ok := r.Config.Profiles["embed"]; ok {
		if ep, ok := r.Adapters[providerID].(EmbedProvider); ok {
			return ep, true
		}
	}
	for _, adapter := range r.Adapters {
		if ep, ok := adapter.(EmbedProvider); ok {
			return ep, true
		}
	}
	return nil, false
}
