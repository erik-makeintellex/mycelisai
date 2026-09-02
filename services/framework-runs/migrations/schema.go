package migrations

import _ "embed"

// CurrentSchema is the only fresh-install schema for the independent Runs database.
//
//go:embed 001_current_schema.sql
var CurrentSchema string
