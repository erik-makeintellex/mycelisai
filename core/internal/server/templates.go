package server

import "time"

// CE-1: Template Engine - Confirm Tokens, Intent Proofs, Audit

const confirmTokenTTL = 15 * time.Minute

// Token validation errors

type tokenError string

func (e tokenError) Error() string { return string(e) }

const (
	errDBUnavailable tokenError = "database not available"
	errInvalidToken  tokenError = "invalid token format"
	errTokenNotFound tokenError = "token not found"
	errTokenConsumed tokenError = "token already consumed"
	errTokenExpired  tokenError = "token expired"
)
