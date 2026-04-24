//go:build darwin

package mcputils

import (
	"os"
	"os/exec"
	"strings"
)

// getFromKeychain retrieves the upstream token from macOS Keychain.
//
// To store a token in Keychain:
//
//	security add-generic-password -s mcpgen-upstream -a mcpgen-upstream -w <your-token>
//
// Customize the service name:
//
//	export MCP_UPSTREAM_KEYCHAIN_SERVICE=mcpgen-upstream
func getFromKeychain() string {
	service := "mcpgen-upstream"
	if s := os.Getenv("MCP_UPSTREAM_KEYCHAIN_SERVICE"); s != "" {
		service = s
	}
	cmd := exec.Command("security", "find-generic-password", "-s", service, "-wa", "")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getFromWinCred() string { return "" }
