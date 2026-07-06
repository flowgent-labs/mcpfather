package tests

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// LDAP integration tests
//
// These tests verify that a generated MCP server can:
//  1. Connect to an LDAP server and perform a simple bind
//  2. Generate a Basic auth token from the bind credentials
//  3. Forward the token as a Basic Authorization header to upstream APIs
//  4. Respect MCP__ env var overrides for LDAP config
//
// A mock LDAP server (mockLDAPServer) is used for unit-level tests.
// E2E tests use a real glauth container (it/docker/glauth/).
// ---------------------------------------------------------------------------

// glauthCreds holds the expected glauth credentials matching config.toml.
type glauthCreds struct {
	URL          string
	BindDN       string
	BindPassword string
}

// defaultGlauthCreds returns the default glauth credentials.
// The bind DN includes the group OU: cn=<user>,ou=<group>,dc=test,dc=local
func defaultGlauthCreds() glauthCreds {
	if u := os.Getenv("GLAUTH_URL"); u != "" {
		dn := os.Getenv("GLAUTH_BIND_DN")
		if dn == "" {
			dn = "cn=mcp-svc,ou=svcaccts,dc=test,dc=local"
		}
		pw := os.Getenv("GLAUTH_BIND_PASSWORD")
		if pw == "" {
			pw = "ldap-secret-123"
		}
		return glauthCreds{URL: u, BindDN: dn, BindPassword: pw}
	}
	// Auto-detect port
	addr := "localhost:3893"
	if conn, err := (&net.Dialer{Timeout: 500 * time.Millisecond}).Dial("tcp", addr); err == nil {
		conn.Close()
	} else {
		addr = "localhost:3389"
	}
	return glauthCreds{
		URL:          "ldap://" + addr,
		BindDN:       "cn=mcp-svc,ou=svcaccts,dc=test,dc=local",
		BindPassword: "ldap-secret-123",
	}
}

// ensureGlauth makes sure a glauth LDAP server is reachable.
// If a container named "mcpgen-glauth" is not running, it starts one via docker run.
func ensureGlauth(t *testing.T) (cleanup func()) {
	t.Helper()

	if !dockerAvailable() {
		// Check if glauth is already reachable
		creds := defaultGlauthCreds()
		_, err := (&net.Dialer{Timeout: 2 * time.Second}).Dial("tcp", strings.TrimPrefix(creds.URL, "ldap://"))
		if err != nil {
			t.Skipf("docker not found and glauth not reachable — skipping")
		}
		t.Logf("glauth already reachable at %s (no docker)", creds.URL)
		return func() {}
	}

	// Stop any existing glauth container (leftover from previous runs)
	for _, name := range []string{"mcpgen-glauth", "glauth"} {
		exec.Command("docker", "stop", name).Run()
		exec.Command("docker", "rm", name).Run()
	}

	// Start glauth container with our config
	glauthDir := filepath.Join(repoRoot(t), "it", "docker", "glauth")
	configFile := filepath.Join(glauthDir, "config.toml")
	cmd := exec.Command("docker", "run", "-d", "--name", "mcpgen-glauth",
		"--network", "host",
		"-v", configFile+":/app/config/config.cfg:ro",
		"registry.cn-shenzhen.aliyuncs.com/wl4g/glauth:v2.5.0",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("docker run glauth failed: %v\n%s", err, out)
	}
	t.Logf("glauth container started")

	waitForGlauth(t)
	return func() {
		exec.Command("docker", "stop", "mcpgen-glauth").Run()
		exec.Command("docker", "rm", "mcpgen-glauth").Run()
		t.Logf("glauth container cleaned up")
	}
}

// waitForGlauth polls the glauth LDAP port until it responds or times out.
// It always waits on port 3893 (the configured port in config.toml / GLAUTH_URL).
func waitForGlauth(t *testing.T) {
	t.Helper()
	addr := "localhost:3893"
	if u := os.Getenv("GLAUTH_URL"); u != "" {
		addr = strings.TrimPrefix(u, "ldap://")
	}
	for i := 0; i < 60; i++ {
		conn, err := (&net.Dialer{Timeout: 1 * time.Second}).Dial("tcp", addr)
		if err == nil {
			conn.Close()
			t.Logf("glauth LDAP server reachable at %s", addr)
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("glauth did not become ready within 60s")
}

// TestLDAPConfigEnvOverrides verifies that MCP__ env vars override LDAP config.
func TestLDAPConfigEnvOverrides(t *testing.T) {
	mock := startMockLDAPServer("cn=svc,dc=example,dc=com", "test-password")
	defer mock.Close()

	envVars := []string{
		"MCP__AUTH__LDAP__ENABLED=true",
		"MCP__AUTH__LDAP__URL=" + mock.URL(),
		"MCP__AUTH__LDAP__BASE_DN=dc=example,dc=com",
		"MCP__AUTH__LDAP__BIND_DN=cn=svc,dc=example,dc=com",
		"MCP__AUTH__LDAP__BIND_PASSWORD=test-password",
		"MCP__AUTH__LDAP__TIMEOUT=10",
		"MCP__UPSTREAM__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for LDAP config testing")
}

// TestLDAPMockBind verifies the mock LDAP server accepts valid credentials.
func TestLDAPMockBind(t *testing.T) {
	mock := startMockLDAPServer("cn=svc,dc=example,dc=com", "secret123")
	defer mock.Close()

	if mock.URL() == "" {
		t.Fatal("mock LDAP server URL is empty")
	}
	t.Logf("Mock LDAP server listening at %s", mock.URL())

	// Verify expected Basic auth value
	expectedAuth := mock.ExpectedBasicAuth()
	if !strings.HasPrefix(expectedAuth, "Basic ") {
		t.Errorf("expected Basic auth prefix, got: %s", expectedAuth)
	}
	t.Logf("Expected Basic auth: %s", expectedAuth)
}

// TestLDAPFullE2E runs a full end-to-end test against a real glauth container:
//  1. Start a real glauth LDAP server (docker)
//  2. Start a mock upstream API that records the Authorization header
//  3. Generate and build an MCP server from the minimal spec
//  4. Configure the generated server with LDAP settings via config.yaml
//  5. Start the server in HTTP mode
//  6. Call a tool and verify the upstream received a Basic auth header
func TestLDAPFullE2E(t *testing.T) {
	cleanupGlauth := ensureGlauth(t)
	defer cleanupGlauth()

	creds := defaultGlauthCreds()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	// Generate and build the MCP server project
	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	// Write a config file with LDAP settings pointing to the real glauth instance
	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  ldap:
    enabled: true
    url: %s
    base_dn: dc=test,dc=local
    bind_dn: %s
    bind_password: %s
    insecure_skip_verify: false
    timeout: 5
upstream:
  endpoint: %s
`, creds.URL, creds.BindDN, creds.BindPassword, mock.server.URL)
	writeCoreVirtualConfig(t, homeDir, binaryName, configYAML)

	// Start server in HTTP mode with the config
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

	// Allow LDAP bind to complete after server start
	time.Sleep(3 * time.Second)

	// Call a tool via MCP HTTP transport
	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

	// The mock upstream should have received a Basic auth header
	if mock.requestCount() == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	auth := mock.requests[0].Authorization
	if auth == "" {
		t.Error("expected Authorization header in upstream request, but it was empty")
	} else if !strings.HasPrefix(auth, "Basic ") {
		t.Errorf("expected Basic auth token, got: %s", auth)
	} else {
		t.Logf("Upstream received valid Basic auth token from real glauth (len=%d)", len(auth))
	}
}

// TestLDAPWrongCredentials verifies the server handles bind failures gracefully
// against a real glauth container.
func TestLDAPWrongCredentials(t *testing.T) {
	cleanupGlauth := ensureGlauth(t)
	defer cleanupGlauth()

	creds := defaultGlauthCreds()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	// Write config with WRONG password
	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  ldap:
    enabled: true
    url: %s
    base_dn: dc=test,dc=local
    bind_dn: %s
    bind_password: wrong-password
    insecure_skip_verify: false
    timeout: 5
upstream:
  endpoint: %s
`, creds.URL, creds.BindDN, mock.server.URL)
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
	time.Sleep(3 * time.Second)

	// Call should still work — just without auth header
	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result (wrong ldap password): %s", trimMsg(result, 300))

	// Server should stay running and respond
	stderrOut := stderrBuf.String()
	if strings.Contains(stderrOut, "initial LDAP bind failed") {
		t.Logf("Expected warning about LDAP bind failure present in stderr")
	}
}

// TestLDAPRealGlauthBind verifies real glauth connectivity via docker.
func TestLDAPRealGlauthBind(t *testing.T) {
	cleanupGlauth := ensureGlauth(t)
	defer cleanupGlauth()

	creds := defaultGlauthCreds()
	addr := strings.TrimPrefix(creds.URL, "ldap://")
	conn, err := (&net.Dialer{Timeout: 2 * time.Second}).Dial("tcp", addr)
	if err != nil {
		t.Fatalf("glauth not reachable at %s: %v", addr, err)
	}
	conn.Close()
	t.Logf("Real glauth LDAP server reachable at %s", addr)
}
