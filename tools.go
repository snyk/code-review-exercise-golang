//go:build tools

package main

// Manage tool dependencies via go.mod.
import (
	_ "golang.org/x/vuln/cmd/govulncheck"
)
