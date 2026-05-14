//go:build tools
// +build tools

package main

import (
	_ "github.com/aws/aws-sdk-go-v2/service/translate"
	_ "github.com/gomarkdown/markdown"
	_ "github.com/gomarkdown/markdown/html"
	_ "github.com/tdewolff/minify/v2"
	_ "github.com/tdewolff/minify/v2/html"
)
