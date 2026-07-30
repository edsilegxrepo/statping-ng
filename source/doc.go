// Package source holds all the assets for Statping. This includes
// CSS, JS, SCSS, HTML and other website related content.
// This package uses Go's native embed directive to compile all assets into the binary.
//
// # Required Dependencies
//
// - sass -> https://sass-lang.com/install
//
// # Compile Assets
//
// The frontend assets are built with Vite and embedded using //go:embed.
// No external tools like rice are required - just build the Go binary:
//
//	cd frontend && npm run build
//	go build -o statping ./cmd
//
// More info on: https://github.com/adamboutcher/statping-ng
package source
