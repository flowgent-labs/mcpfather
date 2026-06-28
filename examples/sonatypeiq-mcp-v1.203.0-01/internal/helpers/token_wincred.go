//go:build windows

package mcputils

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// getFromWinCred retrieves the upstream token from Windows Credential Manager.
// Returns "" on any error, timeout, or panic — this is purely optional.
//
// To store a token:
//
//	cmdkey /add:mcpgen-upstream /user:mcpgen-upstream /pass:your-token
//
// Customize the target name:
//
//	set MCP_UPSTREAM_WINCRED_TARGET=mcpgen-upstream
func getFromWinCred() (token string) {
	defer func() { recover() }()
	target := "mcpgen-upstream"
	if t := os.Getenv("MCP_UPSTREAM_WINCRED_TARGET"); t != "" {
		target = t
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cmdkey", "/get:"+target)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Password:") {
			return strings.TrimPrefix(line, "Password:")
		}
	}
	return ""
}

func getFromKeychain() string { return "" }
