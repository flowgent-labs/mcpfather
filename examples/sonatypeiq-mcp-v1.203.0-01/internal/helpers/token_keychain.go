//go:build darwin

package mcputils

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// getFromKeychain retrieves the upstream token from macOS Keychain.
// Returns "" on any error, timeout, or panic — this is purely optional.
//
// To store a token in Keychain:
//
//	security add-generic-password -s mcpgen-upstream -a mcpgen-upstream -w <your-token>
//
// Customize the service name:
//
//	export MCP_UPSTREAM_KEYCHAIN_SERVICE=mcpgen-upstream
func getFromKeychain() (token string) {
	defer func() { recover() }()
	service := "mcpgen-upstream"
	if s := os.Getenv("MCP_UPSTREAM_KEYCHAIN_SERVICE"); s != "" {
		service = s
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "security", "find-generic-password", "-s", service, "-wa", "")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getFromWinCred() string { return "" }
