package tools

import (
	"context"
	"fmt"

	"github.com/luisra51/mcp-holded/holded"
)

// ensureToolAllowed combines two guardrails:
//  1. The tool is in the allowlist (if any).
//  2. If the tool is a write tool, HOLDED_ALLOW_WRITE must be true.
func ensureToolAllowed(ctx context.Context, toolName string, write bool) error {
	cfg := holded.ConfigFromContext(ctx)
	if !holded.IsToolAllowed(cfg, toolName) {
		return fmt.Errorf("tool not allowed: %s", toolName)
	}
	if write && !cfg.AllowWrite {
		return holded.ErrWriteDisabled
	}
	return nil
}
