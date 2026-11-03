// Package migrations embeds the SQL migration files so the binary can
// apply them on startup without a separate `migrate` CLI step or any
// assumption that a migrations/ directory exists on disk at runtime.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
