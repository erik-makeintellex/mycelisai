package governance

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"google.golang.org/protobuf/types/known/structpb"
)

// Guard evaluates bus envelopes against the loaded policy. It holds no
// approval state: REQUIRE_APPROVAL only means "do not proceed" (C2-RETIRE);
// real approval is the durable confirm-action/proposal path. Engine.Config is
// swapped only under mu; readers take a snapshot under mu.RLock. A nil Config
// means degraded.
type Guard struct {
	Engine  *Engine
	mu      sync.RWMutex
	applyMu sync.Mutex // serializes ReplacePolicy (persist + swap)
	// ingressProviders are extra global-input provider tokens admitted by
	// ValidateIngress (F16); guarded by mu.
	ingressProviders map[string]struct{}
}

func NewGuard(policyPath string) (*Guard, error) {
	engine, err := NewEngine(policyPath)
	if err != nil {
		return nil, err
	}

	return &Guard{Engine: engine}, nil
}

// Intercept evaluates a message and returns (proceed, action). Only ALLOW
// proceeds; DENY, REQUIRE_APPROVAL and unknown actions do not, and nothing is
// parked for later release.
func (g *Guard) Intercept(msg *pb.MsgEnvelope) (bool, string) {
	// Extract Context
	ctx := make(map[string]interface{})
	if msg.SwarmContext != nil {
		for k, v := range msg.SwarmContext.Fields {
			// Basic type mapping for the naive parser
			if _, ok := v.Kind.(*structpb.Value_NumberValue); ok {
				ctx[k] = v.GetNumberValue()
			} else if _, ok := v.Kind.(*structpb.Value_StringValue); ok {
				ctx[k] = v.GetStringValue()
			} else if _, ok := v.Kind.(*structpb.Value_BoolValue); ok {
				ctx[k] = v.GetBoolValue()
			}
		}
	}

	// Also mix in Event Data for context?
	// For "amount > 50", usually amount is in the event data, not the header context.
	// The policy engine might want both.
	// For this loop, let's also pull data from EventPayload if present.
	if msg.GetEvent() != nil && msg.GetEvent().Data != nil {
		for k, v := range msg.GetEvent().Data.Fields {
			if _, ok := v.Kind.(*structpb.Value_NumberValue); ok {
				ctx[k] = v.GetNumberValue()
			}
		}
	}

	// Extract intent
	intent := ""
	if msg.GetEvent() != nil {
		intent = msg.GetEvent().EventType
	}

	cfg := g.policySnapshot()
	if cfg == nil {
		// Degraded: no loaded policy is authority, so nothing is allowed.
		return false, ActionDeny
	}
	action := (&Engine{Config: cfg}).Evaluate(msg.TeamId, msg.SourceAgentId, intent, ctx)

	switch action {
	case ActionAllow:
		return true, action
	case ActionDeny:
		log.Printf("DENY: Guard blocked: %s from %s", intent, msg.SourceAgentId)
		return false, action
	case ActionRequireApproval:
		// Not proceeding is the whole effect; the caller records the
		// observation. Nothing is parked, so nothing can be released later.
		return false, action
	}
	// Unknown action (validation should make this unreachable): fail closed.
	log.Printf("DENY: Guard saw unknown policy action %q for %s", action, intent)
	return false, ActionDeny
}

// policySnapshot returns the live policy pointer under the read lock. Policy
// configs are never mutated after they are swapped in, so the snapshot is safe
// to evaluate without holding the lock.
func (g *Guard) policySnapshot() *PolicyConfig {
	if g == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Engine == nil {
		return nil
	}
	return g.Engine.Config
}

// MaxIngressBytes caps one global-input message.
const MaxIngressBytes = 1024 * 1024

var ingressProviderToken = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

// ValidIngressProviderToken reports whether p is a single lowercase subject
// token that may name a global-input provider (F16): no dots, wildcards,
// slashes, percent signs, whitespace or upper case.
func ValidIngressProviderToken(p string) bool {
	return ingressProviderToken.MatchString(p)
}

// SetIngressProviders replaces the extra provider tokens ValidateIngress
// admits beside the protocol lanes. Invalid tokens are dropped.
func (g *Guard) SetIngressProviders(providers []string) {
	allowed := make(map[string]struct{}, len(providers))
	for _, p := range providers {
		if ValidIngressProviderToken(p) {
			allowed[p] = struct{}{}
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ingressProviders = allowed
}

// ValidateIngress checks raw NATS messages before they enter the Soma
// processing loop: a size cap and an exact-match subject set (F16). The set is
// the protocol's user lane plus the providers given to SetIngressProviders;
// prefixes, wildcards, empty or deep tokens are never admitted.
func (g *Guard) ValidateIngress(subject string, data []byte) error {
	if len(data) > MaxIngressBytes {
		return fmt.Errorf("payload too large: %d bytes", len(data))
	}
	if subject == protocol.TopicGlobalInputUser {
		return nil
	}
	provider, ok := strings.CutPrefix(subject, strings.TrimSuffix(protocol.TopicGlobalInputFmt, "%s"))
	if ok && ValidIngressProviderToken(provider) && g != nil {
		g.mu.RLock()
		_, allowed := g.ingressProviders[provider]
		g.mu.RUnlock()
		if allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid ingress subject: %q", subject)
}

// GetPolicyConfig returns the current policy configuration (nil when degraded).
func (g *Guard) GetPolicyConfig() *PolicyConfig {
	return g.policySnapshot()
}

// UpdatePolicyConfig replaces the in-memory policy configuration. A non-nil
// config clears the degraded state.
func (g *Guard) UpdatePolicyConfig(cfg *PolicyConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Engine == nil {
		g.Engine = &Engine{}
	}
	g.Engine.Config = cfg
}

// ErrAuditUnavailable is returned by a ReplacePolicy apply callback when the
// requested audit could not be written; nothing is changed.
var ErrAuditUnavailable = errors.New("governance audit unavailable")

// ReplacePolicy serializes replacements under applyMu. The apply callback
// receives the live policy snapshot taken under that lock (so previous_digest
// is never stale, A2b), writes the requested audit and persists the file; the
// in-memory swap happens only when apply succeeds, so a failure leaves the live
// policy unchanged and file and memory cannot diverge.
func (g *Guard) ReplacePolicy(cfg *PolicyConfig, apply func(previous *PolicyConfig) error) error {
	if cfg == nil {
		return fmt.Errorf("policy config is required")
	}
	g.applyMu.Lock()
	defer g.applyMu.Unlock()
	if apply != nil {
		if err := apply(g.policySnapshot()); err != nil {
			return err
		}
	}
	g.UpdatePolicyConfig(cfg)
	return nil
}
