package helpdocs

// Entry is one curated documentation page Soma and the API may read.
type Entry struct {
	Slug        string `json:"slug"`
	Label       string `json:"label"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// Section groups related documentation entries.
type Section struct {
	Section string  `json:"section"`
	Docs    []Entry `json:"docs"`
}

// Manifest is the Core-owned read-only docs surface for Soma.
var Manifest = []Section{
	{
		Section: "Start Here",
		Docs: []Entry{
			{Slug: "user-docs-home", Label: "Use Mycelis", Path: "docs/user/README.md", Description: "Operator-first entry point for Soma, teams, resources, outputs, and recovery workflows"},
			{Slug: "soma-chat", Label: "Using Soma Chat", Path: "docs/user/soma-chat.md", Description: "Concrete Soma prompts, outputs, governed proposals, and recovery"},
			{Slug: "workflow-variants-plan-memory", Label: "Workflow Variants + Plan Memory", Path: "docs/user/workflow-variants-and-plan-memory.md", Description: "Direct Soma, compact teams, multi-lane workflows, and durable plan memory"},
			{Slug: "teams-guide", Label: "Teams", Path: "docs/user/teams.md", Description: "Active team work, compact defaults, broad-ask splitting, and lead-centered workflows"},
			{Slug: "resources-guide", Label: "Outputs And Resources", Path: "docs/user/resources.md", Description: "Output files, capabilities, context, exchange, AI engines, and governed resources"},
		},
	},
	{
		Section: "Trust And Setup",
		Docs: []Entry{
			{Slug: "core-concepts", Label: "Core Concepts", Path: "docs/user/core-concepts.md", Description: "Soma, teams, memory, governance, runs, and trust in operator language"},
			{Slug: "memory-guide", Label: "Memory", Path: "docs/user/memory.md", Description: "Trusted recall, memory lanes, governed context, and continuity rules"},
			{Slug: "governance-trust", Label: "Governance & Trust", Path: "docs/user/governance-trust.md", Description: "Approval posture, risk classes, audit visibility, and trusted-memory precedence"},
			{Slug: "settings-access", Label: "Settings And Access", Path: "docs/user/settings-access.md", Description: "Profile, access posture, auth providers, and connected-tool/search boundaries"},
			{Slug: "system-status-recovery", Label: "System Status & Recovery", Path: "docs/user/system-status-recovery.md", Description: "Health signals, degraded recovery actions, and System Checks workflow"},
			{Slug: "run-timeline", Label: "Run Timeline", Path: "docs/user/run-timeline.md", Description: "Execution timelines, status changes, and run navigation paths"},
		},
	},
	{
		Section: "Contributor And Architecture",
		Docs: []Entry{
			{Slug: "docs-home", Label: "Docs Home", Path: "docs/README.md", Description: "Navigation layer for user, developer, testing, release, and compatibility docs"},
			{Slug: "readme", Label: "Repository Overview", Path: "README.md", Description: "Primary development and command contract"},
			{Slug: "testing", Label: "Testing", Path: "docs/TESTING.md", Description: "Unit, integration, browser, and release validation guidance"},
			{Slug: "api-reference", Label: "API Reference", Path: "docs/API_REFERENCE.md", Description: "Endpoint table with request and response shapes"},
			{Slug: "architecture-index", Label: "Architecture Docs Index", Path: "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md", Description: "Curated active architecture set"},
			{Slug: "mycelis-canonical-prd", Label: "Mycelis Canonical PRD", Path: "docs/architecture-library/MYCELIS_CANONICAL_PRD.md", Description: "Single source for product thesis, UX, runtime architecture, governance, outcomes, capabilities, recovery, MVP scope, P0 delivery, and release gates"},
		},
	},
}

func Flatten(sections []Section) []Entry {
	var entries []Entry
	for _, section := range sections {
		entries = append(entries, section.Docs...)
	}
	return entries
}
