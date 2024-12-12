//go:build tools
// +build tools

package tools

// Manage tool dependencies via go.mod.
import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
