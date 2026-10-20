// Package migrations holds the goose SQL migration scripts embedded into the
// migrate command.
package migrations

import "embed"

// SQL is the embedded goose migration scripts.
//
//go:embed *.sql
var SQL embed.FS
