package server

import (
	"fmt"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// Blocker copy (UX1). Codes and HTTP statuses are the contract; this file owns
// only the words. The user variant is plain language for anyone: it never
// names an API path, env var, URL, permission string, or internal term. The
// admin variant may, because an admin can act on it. Role here selects copy
// only; it never grants or denies anything.

// codeAdminRequired is the generic 403 on an admin-only surface.
const codeAdminRequired = "admin_required"

// codeServiceUnavailable is a 503 where a dependency Core needs is down and
// nothing was changed (UX1).
const codeServiceUnavailable = "service_unavailable"

// MCPS direct MCP tool calls: 404 when {tool} is not a discovered tool of a
// registered server, 403 when a read tool's caller lacks outputs:read.
const (
	codeMCPToolNotFound  = "mcp_tool_not_found"
	codeMCPCallForbidden = "mcp_call_forbidden"
)

// MCPA MCP configuration writes: 400 when a request carries env keys the
// library entry does not declare (nothing installed or launched), 404 when a
// delete names no registered server (nothing disconnected or deleted), 502
// when an install registered the server but it could not start.
const (
	codeMCPEnvRejected    = "mcp_env_rejected"
	codeMCPServerNotFound = "mcp_server_not_found"
	codeMCPConnectFailed  = "mcp_connect_failed"
)

// codeTeamServiceOffline marks the agent runtime (Soma's team service) as
// down; it reuses the existing chat transport code.
const codeTeamServiceOffline = "transport_unavailable"

type blockerText struct {
	Message string // envelope error / headline sentence
	Action  string // data.recommended_action
}

type roleBlockerText struct {
	User  blockerText
	Admin blockerText // zero value: admins read the user copy
}

// viewerIsAdmin reports whether the viewer should read admin copy.
func viewerIsAdmin(r *http.Request) bool {
	if r == nil {
		return false
	}
	identity := IdentityFromContext(r.Context())
	return identity != nil && identity.Role == "admin"
}

func (c roleBlockerText) forViewer(admin bool) blockerText {
	if !admin {
		return c.User
	}
	out := c.User
	if c.Admin.Message != "" {
		out.Message = c.Admin.Message
	}
	if c.Admin.Action != "" {
		out.Action = c.Admin.Action
	}
	return out
}

const nothingRan = " Nothing ran and the proposal is still open."

// blockerCopies holds the role-aware copy for normalized blocker codes.
var blockerCopies = map[string]roleBlockerText{
	codeAdminRequired: {
		User: blockerText{"This area is for admins.", "Ask an admin if you need something here."},
		Admin: blockerText{"Your admin account is missing a permission this needs.",
			"Ask a root admin to grant the permission named in required_scope to your account."},
	},
	codeTokenAlreadyUsed: {
		User: blockerText{"This proposal was already approved.",
			"Refresh the conversation to see the result. It will not run twice."},
	},
	codeConfirmerNotProposer: {
		User: blockerText{"Only the person who asked for this, or an admin, can approve it.",
			"Ask them to approve it, or ask an admin." + nothingRan},
		Admin: blockerText{Action: "Approving someone else's proposal needs approvals:decide on your admin account." + nothingRan},
	},
	codeTokenPurposeUnknown: {
		User: blockerText{"This proposal is out of date.", "Ask Soma to propose it again. Nothing ran."},
	},
	codeTokenWrongPurpose: {
		User: blockerText{"This proposal can't be approved here.", "Ask Soma to propose it again. Nothing ran."},
	},
	codeBlueprintMismatch: {
		User: blockerText{"The team plan changed after Soma proposed it.",
			"Ask Soma to update the plan, or launch the plan it proposed without changes. Nothing launched."},
	},
	codeInvalidConfirmToken: {
		User: blockerText{"This proposal is no longer valid.", "Ask Soma to propose it again. Nothing ran."},
	},
	codeCouncilTemplateUnresolved: {
		User: blockerText{"Start this outcome template from Soma in its organization.",
			"Open Soma in the organization that owns this outcome template and ask again. Nothing ran."},
	},
	codeTeamServiceOffline: {
		User: blockerText{"Soma's team service isn't running, so Soma and its teams can't answer right now.",
			"Try again in a moment. If it keeps happening, ask an admin to restart Soma's team service."},
		Admin: blockerText{Action: "Start the agent runtime (NATS and the Soma team service) from System Status or by restarting Core, then try again."},
	},
	codeServiceUnavailable: {
		User: blockerText{"Changes are paused because the activity log is unavailable. Nothing was changed.",
			"Try again in a moment. If it keeps happening, ask an admin to restore the activity log."},
		Admin: blockerText{Action: "Restore the audit store (the Core database) so changes can be recorded, then try again. Nothing was changed."},
	},
	codeMCPToolNotFound: {
		User: blockerText{"That tool isn't available on this connection. Nothing ran.",
			"Refresh the tool list and pick one of the tools it shows."},
		Admin: blockerText{"This MCP server is not registered, or it does not list a tool with this exact name. Nothing ran.",
			"Check the server's discovered tools in Resources and use the exact, case-sensitive tool name."},
	},
	codeMCPCallForbidden: {
		User: blockerText{"Your account can't use this tool directly. Nothing ran.",
			"Ask Soma to do it for you, or ask an admin for access."},
		Admin: blockerText{Action: "Direct read tools need outputs:read on the caller's account. Nothing ran."},
	},
	codeMCPEnvRejected: {
		User: blockerText{"This tool can't be set up with those settings.",
			"Remove the settings this tool doesn't use and try again. Nothing was installed."},
		Admin: blockerText{"This install sets environment variables the library entry does not declare.",
			"Send only the keys in allowed_env_keys (exact spelling and case); rejected_env_keys lists the others. Nothing was installed or started."},
	},
	codeMCPServerNotFound: {
		User: blockerText{"That connection no longer exists. Nothing was changed.",
			"Refresh the list of connected tools and try again."},
		Admin: blockerText{"No MCP server with this id is registered. Nothing was disconnected or deleted.",
			"Refresh Resources to see the registered MCP servers."},
	},
	codeMCPConnectFailed: {
		User: blockerText{"This tool was added but couldn't start, so it isn't available yet.",
			"Ask an admin to check its settings, then try again or remove it."},
		Admin: blockerText{"The MCP server was registered but could not be started or connected; it is saved with status error.",
			"Read detail for the redacted reason, fix the settings and install again, or delete the server."},
	},
	governancePolicyUnavailableCode: {
		// GET /governance/policy is admin-only; the user variant exists so a
		// future non-admin surface cannot leak the admin text.
		User: blockerText{"Safety rules are unavailable.", "Ask an admin to fix the safety rules."},
		Admin: blockerText{"Your saved safety rules couldn't be read.",
			"Fix core/config/policy.yaml and restart Core, or save a valid policy (PUT /api/v1/governance/policy) to unlock without a restart."},
	},
}

// respondBlocker writes a normalized blocker envelope with the table copy for
// code. adminDetail is technical text attached as data.detail for admins only.
func respondBlocker(w http.ResponseWriter, r *http.Request, status int, code string, adminDetail string, extra map[string]string) {
	respondBlockerText(w, r, status, code, blockerCopies[code], adminDetail, extra)
}

// respondBlockerText is respondBlocker with explicit copy; extra keys are
// added to data for every viewer.
func respondBlockerText(w http.ResponseWriter, r *http.Request, status int, code string, copy roleBlockerText, adminDetail string, extra map[string]string) {
	admin := viewerIsAdmin(r)
	text := copy.forViewer(admin)
	data := map[string]string{"code": code}
	if text.Action != "" {
		data["recommended_action"] = text.Action
	}
	if adminDetail != "" && admin {
		data["detail"] = adminDetail
	}
	for k, v := range extra {
		data[k] = v
	}
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: text.Message, Data: data})
}

// missingTeamPlanApproval is POST /intent/commit without a confirm token.
var missingTeamPlanApproval = roleBlockerText{
	User: blockerText{"This team plan needs Soma's proposal before it can launch.",
		"Ask Soma to propose the plan, then approve it. Nothing launched."},
	Admin: blockerText{Action: "Ask Soma to propose the plan, then launch it with the confirm_token from that proposal. Nothing launched."},
}

// mcpDirectCallNeedsApprover is the 403 admin_required on a high-risk direct
// MCP tool call (MCPS D4): only approvers call these tools directly.
var mcpDirectCallNeedsApprover = roleBlockerText{
	User: blockerText{"Only an admin can use this tool directly. Nothing ran.",
		"Ask Soma to propose it. An admin approves the proposal before it runs."},
	Admin: blockerText{"Your admin account is missing a permission this tool needs. Nothing ran.",
		"Ask Soma to propose it, or ask a root admin to grant the permission named in required_scope to your account."},
}

// mcpServerOfflineCopy is the 503 service_unavailable when the MCP server,
// its tool list or the connection pool can't be reached (MCPS D2).
var mcpServerOfflineCopy = roleBlockerText{
	User: blockerText{"That tool's connection is offline right now. Nothing ran.",
		"Try again in a moment. If it keeps happening, ask an admin to reconnect it."},
	Admin: blockerText{Action: "Reconnect the MCP server from Resources, or check the Core database, then try again. Nothing ran."},
}

// approverRequiredCopy explains a tier-2 block by its approverTier reason.
// An admin only sees it when their account lacks approvals:decide.
func approverRequiredCopy(reason string, admin bool) (whatFailed, nextStep string) {
	because := map[string]string{
		"policy":          ", because it falls under your organization's rules",
		"capability_risk": ", because it uses a high-risk tool",
		"cost":            fmt.Sprintf(", because it may cost more than %.2f", approverCostCeiling),
	}[reason]
	whatFailed = "An admin must approve this before Soma starts" + because + "."
	nextStep = "Keep this proposal open for an admin to approve. Nothing has run."
	if admin {
		nextStep = "Your admin account needs approvals:decide to approve this. Nothing has run and the proposal is still open."
	}
	return whatFailed, nextStep
}

// governanceLockDetail is the services/status governance row detail while
// the safety rules are not loaded. The code prefix is stable.
func governanceLockDetail(admin bool) string {
	text := blockerCopies[governancePolicyUnavailableCode].forViewer(admin)
	if !admin {
		return governance.PolicyUnavailableCode + ": Safety rules are unavailable, so most actions are paused. An admin must fix them."
	}
	return governance.PolicyUnavailableCode + ": Safety rules didn't load, so everything except health checks is paused. " + text.Action
}

// availabilityForViewer returns availability with the admin remedy as the
// recommended action for admins; everyone else keeps the user-safe action.
func availabilityForViewer(r *http.Request, availability cognitive.ExecutionAvailability) cognitive.ExecutionAvailability {
	if viewerIsAdmin(r) && availability.AdminAction != "" {
		availability.RecommendedAction = availability.AdminAction
	}
	availability.AdminAction = ""
	return availability
}
