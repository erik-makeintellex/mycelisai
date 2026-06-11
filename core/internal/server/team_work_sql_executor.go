package server

import (
	"context"
	"database/sql"
)

type teamWorkSQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
