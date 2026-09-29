package router

import (
	"log"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// PolicyApprovalObservedAction is the audit action recorded when a bus event
// matches a REQUIRE_APPROVAL policy rule (C2-RETIRE). The event was already
// delivered by NATS, so Core records that it saw it and does not react; there
// is no queue and nothing to approve. Real approval is the durable
// confirm-action and proposal path.
const PolicyApprovalObservedAction = "policy_approval_required_observed"

// ApprovalObservation is what the Router knows about a REQUIRE_APPROVAL bus
// event. Every field comes from the (untrusted) envelope and subject; none of
// it grants anything.
type ApprovalObservation struct {
	Subject       string
	MessageID     string
	TeamID        string
	SourceAgentID string
	Intent        string
}

// ApprovalAuditor durably records one observation and returns the audit id.
// An error or an empty id means the audit store is unavailable.
type ApprovalAuditor func(ApprovalObservation) (string, error)

// SetApprovalAuditor installs the durable audit sink for REQUIRE_APPROVAL
// observations. Until one is set, observations are logged as audit
// unavailable. Safe to call while the Router is running.
func (r *Router) SetApprovalAuditor(a ApprovalAuditor) {
	if a == nil {
		r.auditor.Store(nil)
		return
	}
	r.auditor.Store(&a)
}

// observeApprovalRequired writes exactly one audit record for a
// REQUIRE_APPROVAL bus event, or one structured log line when audit is down.
// It never publishes and never stores the envelope.
func (r *Router) observeApprovalRequired(subject string, env *pb.MsgEnvelope) {
	obs := ApprovalObservation{Subject: subject, MessageID: env.GetId(), TeamID: env.GetTeamId(), SourceAgentID: env.GetSourceAgentId()}
	if ev := env.GetEvent(); ev != nil {
		obs.Intent = ev.GetEventType()
	}
	var auditID string
	var err error
	if a := r.auditor.Load(); a != nil {
		auditID, err = (*a)(obs)
	}
	if err != nil || auditID == "" {
		log.Printf("governance: %s audit_unavailable=true subject=%q team=%q agent=%q intent=%q message_id=%q err=%v",
			PolicyApprovalObservedAction, obs.Subject, obs.TeamID, obs.SourceAgentID, obs.Intent, obs.MessageID, err)
		return
	}
	log.Printf("governance: %s audit_id=%s subject=%q team=%q agent=%q intent=%q",
		PolicyApprovalObservedAction, auditID, obs.Subject, obs.TeamID, obs.SourceAgentID, obs.Intent)
}
