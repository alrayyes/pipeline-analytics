// Package webassets embeds the built SvelteKit frontend into the Go binary.
// web/'s vite.config.ts points adapter-static's output straight at dist/
// here (go:embed can't reach outside its own package's directory), so
// `bun run build` in web/ must run before this package's dist/ has the
// real frontend in it -- see dist/.gitkeep for what's there otherwise.
package webassets

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS returns the embedded frontend's static files, rooted at dist/ so
// paths read e.g. "index.html" rather than "dist/index.html".
func FS() (fs.FS, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("root embedded frontend at dist: %w", err)
	}

	return sub, nil
}
