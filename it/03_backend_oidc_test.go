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
	_, err := exec.LookPath("docker")
	return err == nil
}

// ---------------------------------------------------------------------------
// Real Keycloak provider — backend OIDC tests
// ---------------------------------------------------------------------------

// ensureBackendKeycloak makes sure a Keycloak OIDC provider is reachable.
func ensureBackendKeycloak(t *testing.T) (cleanup func()) {
	t.Helper()

	if !dockerAvailable() {
		resp, err := http.Get("http://127.0.0.1:8080/realms/master/.well-known/openid-configuration")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Skipf("docker not found and Keycloak not reachable -- skipping")
		}
		if resp != nil {
			resp.Body.Close()
		}
		t.Logf("Keycloak already reachable (no docker)")
		return func() {}
	}

	for _, name := range []string{"mcpfather-keycloak", "keycloak"} {
		exec.Command("docker", "stop", name).Run()
		exec.Command("docker", "rm", name).Run()
	}

	cmd := exec.Command("docker", "run", "-d", "--name", "mcpfather-keycloak",
		"--network", "host",
		"--hostname", "127.0.0.1",
		"-e", "KC_BOOTSTRAP_ADMIN_USERNAME=admin",
		"-e", "KC_BOOTSTRAP_ADMIN_PASSWORD=admin",
		"registry.cn-shenzhen.aliyuncs.com/wl4g/keycloak:26.7.0",
		"start-dev", "--hostname=127.0.0.1", "--hostname-strict=false",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker run keycloak failed: %v\n%s -- skipping OIDC integration test", err, out)
	}
	t.Logf("Keycloak container started")

	waitForKeycloak(t)
	return func() {
		exec.Command("docker", "stop", "mcpfather-keycloak").Run()
		exec.Command("docker", "rm", "mcpfather-keycloak").Run()
	}
}

func waitForKeycloak(t *testing.T) {
	t.Helper()
	discoveryURL := "http://127.0.0.1:8080/realms/master/.well-known/openid-configuration"
	for i := 0; i < 120; i++ {
		resp, err := http.Get(discoveryURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("Keycloak OIDC discovery ready at %s", discoveryURL)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("Keycloak did not become ready within 120s -- skipping OIDC integration test")
}

// ---------------------------------------------------------------------------
// Backend OIDC: Keycloak discovery & config tests
// ---------------------------------------------------------------------------

// TestOIDCConfigEnvOverrides verifies that MCP__ env vars override OIDC config values.
func TestOIDCConfigEnvOverrides(t *testing.T) {
	cleanup := ensureBackendKeycloak(t)
	defer cleanup()

	envVars := []string{
		"MCP__AUTH__BACKEND__OIDC__ENABLED=true",
		"MCP__AUTH__BACKEND__OIDC__ISSUER=http://127.0.0.1:8080/realms/master",
		"MCP__AUTH__BACKEND__OIDC__CLIENT_ID=mcpfather-client",
		"MCP__AUTH__BACKEND__OIDC__CLIENT_SECRET=mcpfather-secret",
		"MCP__AUTH__BACKEND__OIDC__SCOPES=openid",
		"MCP__UPSTREAM__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for OIDC config testing")
}

// TestOIDCKeycloakDiscovery verifies OIDC discovery against a real Keycloak instance.
func TestOIDCKeycloakDiscovery(t *testing.T) {
	cleanup := ensureBackendKeycloak(t)
	defer cleanup()

	resp, err := http.Get("http://127.0.0.1:8080/realms/master/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("Keycloak discovery request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Keycloak discovery returned %d", resp.StatusCode)
	}

	t.Logf("Real Keycloak OIDC discovery OK at http://127.0.0.1:8080/realms/master")
}

// ---------------------------------------------------------------------------
// Mock OIDC provider — custom-claims JWT crafting for negative tests
//
// This server exists ONLY for negative-test token crafting via /sign
// (expired, wrong audience, algorithm confusion, etc.).
// Real OIDC flows (discovery, connectivity, client_credentials, device_code)
// are tested against a real Keycloak container (ensureBackendKeycloak / ensureKeycloak).
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
// Prefer ensureBackendKeycloak (real Keycloak container) for standard OIDC flows.
// This function is retained only for:
//   - negative-test token crafting via /sign (expired, wrong audience, etc.)
func startMockOIDCServer(t *testing.T) *mockOIDCServer {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "mockoidcsvc")
	srcDir := filepath.Join(repoRoot(t), "it", "cmd", "mockoidcsvc")
	buildCmd := exec.Command("go", "build", "-o", binPath, srcDir)
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

func (p *mockOIDCServer) Close()                        { p.cancel(); p.cmd.Wait() }
func (p *mockOIDCServer) Issuer() string                 { return p.issuer }
func (p *mockOIDCServer) TokenURL() string               { return p.issuer + "/token" }
func (p *mockOIDCServer) Audience() string               { return "mcpfather" }

func (p *mockOIDCServer) Token(t *testing.T) string {
	t.Helper()
	body := strings.NewReader("grant_type=client_credentials&client_id=mcpfather-client&client_secret=mcpfather-secret")
	resp, err := http.Post(p.TokenURL(), "application/x-www-form-urlencoded", body)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("token response missing access_token")
	}
	return result.AccessToken
}

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
// Backend OIDC: token exchange & E2E tests (mock provider)
// ---------------------------------------------------------------------------

// TestOIDCTokenExchange verifies the OIDC provider issues tokens via client_credentials.
func TestOIDCTokenExchange(t *testing.T) {
	oidc := startMockOIDCServer(t)
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

// TestOIDCFullE2E runs a full end-to-end test with the mock OIDC provider.
func TestOIDCFullE2E(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  backend:
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
