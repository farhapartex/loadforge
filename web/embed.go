package web

import "embed"

//go:embed static templates
var embeddedFS embed.FS
