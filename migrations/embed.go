// Package migrations embeds the .sql files next to it so the compiled binary
// can migrate its own database. There is no separate migration tool to install.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
