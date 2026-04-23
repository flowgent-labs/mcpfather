//go:build windows

package mcputils

import (
	"os"
	"os/exec"
	"strings"
)

// getFromWinCred retrieves the upstream token from Windows Credential Manager.
//
// To store a token:
//
//	cmdkey /add:mcpgen-upstream /user:mcpgen-upstream /pass:your-token
//
// Customize the target name:
//
//	set MCP_UPSTREAM_WINCRED_TARGET=mcpgen-upstream
func getFromWinCred() string {
	target := "mcpgen-upstream"
	if t := os.Getenv("MCP_UPSTREAM_WINCRED_TARGET"); t != "" {
		target = t
	}
	cmd := exec.Command("cmdkey", "/get:"+target)
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
