//go:build tools

// Package tools pins tool dependencies in go.mod.
package tools

import (
	_ "github.com/vektra/mockery/v2"
	_ "gotest.tools/gotestsum"
)
