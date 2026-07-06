package tests

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// testOIDCProvider wraps a standalone OIDC provider subprocess.
// It is a real, separately-compiled binary that speaks real OIDC protocol
// on a real TCP port — not an in-process httptest.Server.
type testOIDCProvider struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	addr   string // "host:port"
	issuer string // "http://host:port"
}

// startTestOIDCProvider compiles and starts a standalone OIDC provider binary
// that supports the client_credentials grant. The binary is compiled from
// it/cmd/testoidc/main.go and runs as a separate OS process.
func startTestOIDCProvider(t *testing.T) *testOIDCProvider {
	t.Helper()

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "testoidc")
	srcDir := filepath.Join(repoRoot(t), "it", "cmd", "testoidc")
	buildCmd := exec.Command("go", "build", "-o", binPath, srcDir)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build testoidc: %v\n%s", err, out)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, binPath, "-clients", "mcpgen-client:mcpgen-secret,test-client:test-secret")
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start testoidc: %v", err)
	}

	// Read the listen address from the first line of stdout
	ch := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			ch <- ""
			return
		}
		// Drain stdout to background so the subprocess doesn't block
		go io.Copy(io.Discard, reader)
		ch <- line[:len(line)-1] // trim newline
	}()

	var addr string
	select {
	case addr = <-ch:
	case <-time.After(10 * time.Second):
		cancel()
		cmd.Wait()
		t.Fatal("testoidc did not print address within 10s")
	}

	if addr == "" {
		cancel()
		cmd.Wait()
		t.Fatal("testoidc failed to print listen address")
	}

	// Wait for the provider to be ready
	baseURL := "http://" + addr
	for i := 0; i < 50; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Cleanup(func() {
		cancel()
		cmd.Wait()
	})

	return &testOIDCProvider{
		cmd:    cmd,
		cancel: cancel,
		addr:   addr,
		issuer: baseURL,
	}
}

func (p *testOIDCProvider) Close() {
	p.cancel()
	p.cmd.Wait()
}

func (p *testOIDCProvider) Issuer() string  { return p.issuer }
func (p *testOIDCProvider) TokenURL() string { return p.issuer + "/token" }
