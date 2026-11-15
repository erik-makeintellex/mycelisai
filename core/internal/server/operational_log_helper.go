package server

import (
	"log"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// OperationalLogOption mutates an OperationalLogContext before normalization.
type OperationalLogOption func(*protocol.OperationalLogContext)

// WithOperationalRunID attaches a run_id to an OperationalLogContext.
func WithOperationalRunID(runID string) OperationalLogOption {
	return func(ctx *protocol.OperationalLogContext) {
		ctx.RunID = strings.TrimSpace(runID)
	}
}

// WithOperationalTeamID attaches a team_id to an OperationalLogContext.
func WithOperationalTeamID(teamID string) OperationalLogOption {
	return func(ctx *protocol.OperationalLogContext) {
		ctx.TeamID = strings.TrimSpace(teamID)
	}
}

// WithOperationalSource attaches source_kind and source_channel metadata.
func WithOperationalSource(kind protocol.SignalSourceKind, channel string) OperationalLogOption {
	return func(ctx *protocol.OperationalLogContext) {
		ctx.SourceKind = kind
		ctx.SourceChannel = strings.TrimSpace(channel)
	}
}

// WithOperationalWhyItMatters explains operational impact for human and agentry review.
func WithOperationalWhyItMatters(why string) OperationalLogOption {
	return func(ctx *protocol.OperationalLogContext) {
		ctx.WhyItMatters = strings.TrimSpace(why)
	}
}

// BuildOperationalErrorContext creates a normalized V8.2 OperationalLogContext for
// human operators and meta-agentry (Soma, MetaAgentry, TeamLead) to review on failure.
func BuildOperationalErrorContext(
	component string,
	summary string,
	detail string,
	suggestedAction string,
	opts ...OperationalLogOption,
) protocol.OperationalLogContext {
	ctx := protocol.OperationalLogContext{
		SchemaVersion:   protocol.OperationalLogSchemaVersion,
		ReviewScope:     protocol.LogReviewScopeCentralReview,
		Service:         "mycelis-core",
		Component:       strings.TrimSpace(component),
		Summary:         strings.TrimSpace(summary),
		Detail:          strings.TrimSpace(detail),
		SuggestedAction: strings.TrimSpace(suggestedAction),
		SourceKind:      protocol.SourceKindWebAPI,
		PayloadKind:     "operational_error",
		Status:          "error",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&ctx)
		}
	}
	return ctx.Normalize()
}

// LogOperationalError builds a normalized V8.2 OperationalLogContext and emits a
// standardized, component-identified log record for both human operators and agentry.
func LogOperationalError(
	component string,
	summary string,
	detail string,
	suggestedAction string,
	opts ...OperationalLogOption,
) protocol.OperationalLogContext {
	ctx := BuildOperationalErrorContext(component, summary, detail, suggestedAction, opts...)
	log.Printf(
		"[%s] status=%s review_scope=%s summary=%q detail=%q suggested_action=%q run_id=%q team_id=%q review_channels=%s",
		ctx.Component,
		ctx.Status,
		ctx.ReviewScope,
		ctx.Summary,
		ctx.Detail,
		ctx.SuggestedAction,
		ctx.RunID,
		ctx.TeamID,
		strings.Join(ctx.ReviewChannels, ","),
	)
	return ctx
}
