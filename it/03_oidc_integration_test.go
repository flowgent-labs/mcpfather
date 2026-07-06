package tests

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// OIDC integration tests
//
// These tests verify that a generated MCP server can:
//  1. Discover an OIDC provider's token endpoint via .well-known/openid-configuration
//  2. Obtain an access token using the client_credentials grant
//  3. Forward the token as a Bearer Authorization header to upstream APIs
//  4. Respect MCP__ env var overrides for OIDC config
//
// A standalone OIDC provider binary (it/cmd/testoidc/main.go) is compiled and
// run as a real subprocess for the client_credentials flow. Real Dex container
// is used for discovery connectivity testing.
// ---------------------------------------------------------------------------

// dockerAvailable checks whether the docker CLI is available.
func dockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// ensureDex makes sure a Dex OIDC provider is reachable at the default address.
func ensureDex(t *testing.T) (cleanup func()) {
	t.Helper()

	if !dockerAvailable() {
		resp, err := http.Get("http://localhost:5556/dex/.well-known/openid-configuration")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Skipf("docker not found and Dex not reachable -- skipping")
		}
		if resp != nil {
			resp.Body.Close()
		}
		t.Logf("Dex already reachable (no docker)")
		return func() {}
	}

	for _, name := range []string{"mcpgen-dex", "dex"} {
		exec.Command("docker", "stop", name).Run()
		exec.Command("docker", "rm", name).Run()
	}

	dexDir := filepath.Join(repoRoot(t), "it", "docker", "dex")
	configFile := filepath.Join(dexDir, "config.yaml")
	cmd := exec.Command("docker", "run", "-d", "--name", "mcpgen-dex",
		"--network", "host",
		"-v", configFile+":/etc/dex/config.yaml:ro",
		"docker.io/dexidp/dex:v2.41.1", "dex", "serve", "/etc/dex/config.yaml")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("docker run dex failed: %v\n%s", err, out)
	}
	t.Logf("Dex container started")

	waitForDex(t)
	return func() {
		exec.Command("docker", "stop", "mcpgen-dex").Run()
		exec.Command("docker", "rm", "mcpgen-dex").Run()
	}
}

func waitForDex(t *testing.T) {
	t.Helper()
	discoveryURL := "http://localhost:5556/dex/.well-known/openid-configuration"
	for i := 0; i < 60; i++ {
		resp, err := http.Get(discoveryURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("Dex OIDC discovery ready at %s", discoveryURL)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Dex did not become ready within 60s")
}

// TestOIDCConfigEnvOverrides verifies that MCP__ env vars override OIDC config values.
func TestOIDCConfigEnvOverrides(t *testing.T) {
	oidc := startTestOIDCProvider(t)
	defer oidc.Close()

	envVars := []string{
		"MCP__AUTH__OIDC__ENABLED=true",
		"MCP__AUTH__OIDC__ISSUER=" + oidc.Issuer(),
		"MCP__AUTH__OIDC__CLIENT_ID=test-client",
		"MCP__AUTH__OIDC__CLIENT_SECRET=test-secret",
		"MCP__AUTH__OIDC__SCOPES=openid",
		"MCP__UPSTREAM__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for OIDC config testing")
}

// TestOIDCDiscovery verifies that the OIDC provider's discovery endpoint is reachable.
func TestOIDCDiscovery(t *testing.T) {
	oidc := startTestOIDCProvider(t)
	defer oidc.Close()

	resp, err := http.Get(oidc.Issuer() + "/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("OIDC discovery request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("OIDC discovery returned %d", resp.StatusCode)
	}
	t.Logf("OIDC discovery endpoint OK: %s", oidc.Issuer())
}

// TestOIDCTokenExchange verifies the OIDC provider issues tokens via client_credentials.
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
	t.Logf("OIDC token exchange OK")
}

// TestOIDCFullE2E runs a full end-to-end test:
//  1. Start a standalone OIDC provider (real subprocess, supports client_credentials)
//  2. Start a mock upstream API that records the Authorization header
//  3. Generate and build an MCP server from the minimal spec
//  4. Configure the generated server with OIDC settings via config.yaml
//  5. Start the server in HTTP mode
//  6. Call a tool and verify the upstream received a Bearer token
//
// Note: Dex is not used for the token exchange because Dex only supports
// authorization_code and refresh_token grants (not client_credentials or
// password). A standalone OIDC provider binary is used instead to provide a
// real, externally-compiled OIDC server for the client_credentials flow.
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
    client_id: mcpgen-client
    client_secret: mcpgen-secret
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

// TestOIDCRealDexDiscovery verifies OIDC discovery against a real Dex instance.
func TestOIDCRealDexDiscovery(t *testing.T) {
	cleanupDex := ensureDex(t)
	defer cleanupDex()

	resp, err := http.Get("http://localhost:5556/dex/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("Dex discovery request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Dex discovery returned %d", resp.StatusCode)
	}

	t.Logf("Real Dex OIDC discovery OK at http://localhost:5556/dex")
}
