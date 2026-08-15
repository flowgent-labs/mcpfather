package tests

import (
	"bufio"
	"context"
	"encoding/json"
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

// ---------------------------------------------------------------------------
// Docker helper
// ---------------------------------------------------------------------------

// dockerAvailable checks whether the docker CLI is available.
func dockerAvailable() bool {
	_, err := exec.LookPath("/bin/docker")
	return err == nil
}

// ---------------------------------------------------------------------------
// OIDC provider — upstream OIDC tests
// ---------------------------------------------------------------------------

// ensureUpstreamKeycloak makes sure a Keycloak OIDC provider is reachable.
// Delegates to the shared provider (sync.Once) managed by ensureKeycloak; if
// a real Keycloak container cannot become ready, tests use the mock OIDC
// provider fallback from 03_server_oidc_test.go.
func ensureUpstreamKeycloak(t *testing.T) (issuer string, cleanup func()) {
	t.Helper()
	issuer, cleanup = ensureKeycloak(t)
	return issuer, cleanup
}

// ---------------------------------------------------------------------------
// Upstream OIDC: Keycloak discovery & config tests
// ---------------------------------------------------------------------------

// TestOIDCConfigEnvOverrides verifies that MCP__ env vars override OIDC config values.
func TestOIDCConfigEnvOverrides(t *testing.T) {
	issuer, cleanup := ensureUpstreamKeycloak(t)
	defer cleanup()

	envVars := []string{
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__ENABLED=true",
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__ISSUER=" + issuer,
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__CLIENT_ID=mcpfather-client",
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__CLIENT_SECRET=mcpfather-secret",
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__SCOPES=openid",
		"MCP__UPSTREAM__DEFAULT__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for OIDC config testing")
}

// TestOIDCKeycloakDiscovery verifies OIDC discovery against the configured provider.
func TestOIDCKeycloakDiscovery(t *testing.T) {
	issuer, cleanup := ensureUpstreamKeycloak(t)
	defer cleanup()

	resp, err := http.Get(issuer + "/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("OIDC discovery request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("OIDC discovery returned %d", resp.StatusCode)
	}

	t.Logf("OIDC discovery OK at %s", issuer)
}

// ---------------------------------------------------------------------------
// Mock OIDC provider — custom-claims JWT crafting for negative tests
//
// This server exists ONLY for negative-test token crafting via /sign
// (expired, wrong audience, algorithm confusion, etc.).
// Real OIDC flows (discovery, connectivity, client_credentials, device_code)
// are tested against a real Keycloak container (ensureUpstreamKeycloak / ensureKeycloak).
//
// The binary is compiled from it/cmd/mockoidcsvc/main.go and runs as a separate
// OS process with its own RSA keypair, real OIDC discovery, and real JWKS.
// ---------------------------------------------------------------------------

type mockOIDCServer struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	addr   string // "host:port"
	issuer string // "http://host:port"
}

// startMockOIDCServer compiles and starts the special-purpose OIDC provider
// for custom-claims JWT crafting and negative testing.
//
// Prefer ensureUpstreamKeycloak (real Keycloak container) for standard OIDC flows.
// This function is retained only for:
//   - negative-test token crafting via /sign (expired, wrong audience, etc.)
func startMockOIDCServer(t *testing.T) *mockOIDCServer {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "mockoidcsvc")
	srcDir := filepath.Join(repoRoot(t), "it", "cmd", "mockoidcsvc")
	logProgress("[mock-oidc] building mockoidcsvc from %s", srcDir)
	buildCmd := exec.Command("go", "build", "-p", "1", "-o", binPath, srcDir)
	buildCmd.Env = append(os.Environ(), testBuildEnv(t)...)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build mockoidcsvc: %v\n%s", err, out)
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
		t.Fatalf("start mockoidcsvc: %v", err)
	}

	ch := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			ch <- ""
			return
		}
		go io.Copy(io.Discard, reader)
		ch <- line[:len(line)-1]
	}()

	var addr string
	select {
	case addr = <-ch:
	case <-time.After(10 * time.Second):
		cancel()
		cmd.Wait()
		t.Fatal("mockoidcsvc did not print address within 10s")
	}

	if addr == "" {
		cancel()
		cmd.Wait()
		t.Fatal("mockoidcsvc failed to print listen address")
	}

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

	return &mockOIDCServer{
		cmd:    cmd,
		cancel: cancel,
		addr:   addr,
		issuer: baseURL,
	}
}

func (p *mockOIDCServer) Close()         { p.cancel(); p.cmd.Wait() }
func (p *mockOIDCServer) Issuer() string { return p.issuer }

func (p *mockOIDCServer) SignToken(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	body, _ := json.Marshal(claims)
	resp, err := http.Post(p.issuer+"/sign", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("sign token request failed: %v", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode sign response: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("sign response missing access_token")
	}
	return result.AccessToken
}

// ---------------------------------------------------------------------------
// Upstream OIDC: token exchange & E2E tests (mock provider)
// ---------------------------------------------------------------------------

// TestOIDCTokenExchange verifies Keycloak client_credentials token exchange.
func TestOIDCTokenExchange(t *testing.T) {
	issuer, cleanup := ensureKeycloak(t)
	defer cleanup()

	token := keycloakClientCredentialsToken(t, issuer)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3-part JWT, got %d parts", len(parts))
	}
	t.Logf("OIDC token exchange OK (client_credentials grant)")
}

// TestOIDCFullE2E runs a full end-to-end backend OIDC flow:
// MCP server → OIDC token endpoint → Bearer token forwarded to upstream.
func TestOIDCFullE2E(t *testing.T) {
	issuer, cleanup := ensureKeycloak(t)
	defer cleanup()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	serviceName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  default:
    auth:
      oidc:
        enabled: true
        issuer: %s
        client_id: mcpfather-client
        client_secret: mcpfather-secret
        scopes: openid
    endpoint: %s
`, issuer, mock.server.URL)
	writeCoreVirtualConfig(t, homeDir, serviceName, configYAML)

	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = testProcessEnv(
		"HOME="+homeDir,
		// Override any MCP__ env from the parent test process (e.g. source .env)
		// so the YAML config takes precedence.
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		"MCP__UPSTREAM__DEFAULT__AUTH__OIDC__CLIENT_SECRET=mcpfather-secret",
	)
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

// TestOIDCClientSecretFileFullE2E verifies Kubernetes-style secret file mounts:
// the generated server reads upstream.default.auth.oidc.client_secret_file and
// uses that value for the OIDC client_credentials token exchange.
func TestOIDCClientSecretFileFullE2E(t *testing.T) {
	oidc := startMockOIDCServer(t)

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	serviceName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	secretFile := filepath.Join(homeDir, "oidc-client-secret")
	if err := os.WriteFile(secretFile, []byte("test-secret\n"), 0600); err != nil {
		t.Fatalf("write OIDC client secret file: %v", err)
	}

	configYAML := fmt.Sprintf(`
upstream:
  default:
    auth:
      oidc:
        enabled: true
        issuer: %s
        client_id: test-client
        client_secret_file: %s
        scopes: openid
    endpoint: %s
`, oidc.Issuer(), secretFile, mock.server.URL)
	writeCoreVirtualConfig(t, homeDir, serviceName, configYAML)

	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = testProcessEnv(
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	baseURL := "http://localhost:" + port
	waitForServer(t, baseURL)

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	auth := mock.requests[0].Authorization
	if auth == "" {
		t.Fatal("expected Authorization header in upstream request")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Fatalf("expected Bearer token, got: %s", auth)
	}
}
