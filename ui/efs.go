package ui

import (
	"embed"
)

//go:embed "html" "static"
var Files embed.FS // NOTE: the above is a special comment, a comment directive.
