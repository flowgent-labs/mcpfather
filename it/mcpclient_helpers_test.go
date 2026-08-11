package tests

import (
	"bytes"
	"os/exec"
	"testing"
)

func runGeneratedMcpclientSh(t *testing.T, clientSh, endpoint, downloadDir string, args ...string) (stdout string, stderr string) {
	t.Helper()
	cmdArgs := append([]string{clientSh}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Env = testProcessEnv(
		"MCP_SERVER_ENDPOINT="+endpoint,
		"MCP_SERVER_DOWNLOAD_DIR="+downloadDir,
	)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("mcpclient.sh %v failed: %v\nstderr:\n%s\nstdout:\n%s", args, err, stderrBuf.String(), stdoutBuf.String())
	}
	return stdoutBuf.String(), stderrBuf.String()
}
