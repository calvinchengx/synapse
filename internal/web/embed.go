package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Assets returns the embedded frontend filesystem rooted at dist/.
func Assets() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
