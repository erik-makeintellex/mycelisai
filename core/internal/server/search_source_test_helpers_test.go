package server

import "github.com/DATA-DOG/go-sqlmock"

func searchSourceRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "provider", "source_type", "endpoint", "scope_kind", "scope_ref",
		"boundary", "auth_scheme", "secret_ref", "mode", "sensitivity_class",
		"trust_class", "status", "recovery",
	})
}
