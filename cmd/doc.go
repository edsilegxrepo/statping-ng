// Package main for building the Statping CLI binary application. This package
// connects to all the other packages to make a runnable binary for multiple
// operating system.
//
// # Build Frontend Assets
//
// Before building, compile the frontend with Vite:
//
//	cd frontend && npm run build
//
// # Build Statping Binary
//
// To build the statping binary for your local environment, run the command below:
//
//	go build -o statping ./cmd
//
// # Build All Binary Arch's
//
// To build Statping for Mac, Windows, Linux, and ARM devices, you can run xgo to build for all. xgo is an awesome
// golang package that requires Docker. https://github.com/crazy-max/xgo
//
//	docker pull crazy-max/xgo
//	build-all
//
// More info on: https://github.com/adamboutcher/statping-ng
package main
