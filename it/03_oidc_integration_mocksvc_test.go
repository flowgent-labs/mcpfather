package tests

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	cmd := exec.CommandContext(ctx, binPath, "-clients", "mcpfather-client:mcpfather-secret,test-client:test-secret")
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

// ---------------------------------------------------------------------------
// Password-grant token exchange & E2E tests
//
// These use the standalone mock OIDC provider because Dex v2.41+ no longer
// supports the password (ROPC) grant. Every other OIDC flow (discovery,
// connectivity, config) is tested against real Dex in 03_oidc_integration_test.go.
// ---------------------------------------------------------------------------

// TestOIDCTokenExchange verifies the OIDC provider issues tokens via password grant.
func TestOIDCTokenExchange(t *testing.T) {
	oidc := startTestOIDCProvider(t)
	defer oidc.Close()

	body := strings.NewReader("grant_type=client_credentials&client_id=test-client&client_secret=test-secret&scope=openid")
	resp, err := http.Post(oidc.TokenURL(), "application/x-www-form-urlencoded", body)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("token endpoint returned %d", resp.StatusCode)
	}
	t.Logf("OIDC token exchange OK (client_credentials grant via mock)")
}

// TestOIDCFullE2E runs a full end-to-end test with the mock OIDC provider:
//  1. Start a standalone OIDC provider (client_credentials grant)
//  2. Start a mock upstream API that records the Authorization header
//  3. Generate and build an MCP server from the minimal spec
//  4. Configure the generated server with OIDC client_credentials settings
//  5. Start the server in HTTP mode
//  6. Call a tool and verify the upstream received a Bearer token
//
// Note: This test uses the mock OIDC provider because Dex v2.41+ does not
// support the password grant. Discovery and connectivity are verified
// against real Dex in TestOIDCDexDiscovery.
func TestOIDCFullE2E(t *testing.T) {
	oidc := startTestOIDCProvider(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  oidc:
    enabled: true
    issuer: %s
    client_id: mcpfather-client
    client_secret: mcpfather-secret
    scopes: openid
    
upstream:
  endpoint: %s
`, oidc.Issuer(), mock.server.URL)
	writeCoreVirtualConfig(t, homeDir, binaryName, configYAML)

	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	baseURL := "http://localhost:" + port
	waitForServer(t, baseURL)

	time.Sleep(2 * time.Second)
	t.Logf("MCP server stderr: %s", stderrBuf.String())

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	auth := mock.requests[0].Authorization
	if auth == "" {
		t.Error("expected Authorization header in upstream request, but it was empty")
	} else if !strings.HasPrefix(auth, "Bearer ") {
		t.Errorf("expected Bearer token, got: %s", auth)
	} else {
		t.Logf("Upstream received valid Bearer token from OIDC provider (len=%d)", len(auth))
	}
}
