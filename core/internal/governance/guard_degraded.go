package governance

import (
	"log"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// PolicyUnavailableCode is the fixed, non-sensitive load-error code reported
// while the Guard runs without a loaded policy. Raw load errors go to the log.
const PolicyUnavailableCode = "policy_unavailable"

// NewDegradedGuard returns a Guard with no loaded policy. It fails closed:
// Intercept denies every envelope and posture-scoped work requires approval.
// A valid admin policy replacement (UpdatePolicyConfig/ReplacePolicy) clears
// the degraded state without a restart. reason is logged once and never
// exposed through an API.
func NewDegradedGuard(reason error) *Guard {
	log.Printf("WARN: Governance policy unavailable (%v). Core is running with governance DEGRADED: "+
		"the Gatekeeper denies all non-heartbeat traffic and posture work requires approval. "+
		"Fix core/config/policy.yaml and restart Core, or PUT a valid policy as an admin.", reason)
	return &Guard{
		Engine:        &Engine{},
		PendingBuffer: make(map[string]*pb.ApprovalRequest),
	}
}

// Degraded reports whether the Guard has no loaded policy. A nil Guard is
// degraded.
func (g *Guard) Degraded() bool {
	return g.policySnapshot() == nil
}

// LoadError returns PolicyUnavailableCode while degraded, otherwise "".
func (g *Guard) LoadError() string {
	if g.Degraded() {
		return PolicyUnavailableCode
	}
	return ""
}
