package protocol

import "strings"

// Work modes shared by collaboration groups and template delivery defaults.
// A work mode says how far a team or group may act before an operator
// approves. Bootstrap starter validation and group handlers must agree on
// this set, so it lives in protocol.
const (
	WorkModeReadOnly            = "read_only"
	WorkModeProposeOnly         = "propose_only"
	WorkModeExecuteWithApproval = "execute_with_approval"
	WorkModeExecuteBounded      = "execute_bounded"
	WorkModeResumeContinuity    = "resume_continuity"
)

// WorkModes lists every known work mode in display order.
var WorkModes = []string{
	WorkModeReadOnly,
	WorkModeProposeOnly,
	WorkModeExecuteWithApproval,
	WorkModeExecuteBounded,
	WorkModeResumeContinuity,
}

// IsKnownWorkMode reports whether mode (trimmed) is a known work mode.
func IsKnownWorkMode(mode string) bool {
	mode = strings.TrimSpace(mode)
	for _, known := range WorkModes {
		if known == mode {
			return true
		}
	}
	return false
}
