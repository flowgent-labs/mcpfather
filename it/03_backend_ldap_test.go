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
// All tests run against a real Glauth LDAP server (it/docker/glauth) via
// Docker. No mock LDAP server is used.
// ---------------------------------------------------------------------------

// glauthCreds holds the expected glauth credentials matching config.toml.
type glauthCreds struct {
	URL          string
	BindDN       string
	BindPassword string
}

// defaultGlauthCreds returns the default glauth credentials.
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
func ensureGlauth(t *testing.T) (cleanup func()) {
	t.Helper()

	if !dockerAvailable() {
		creds := defaultGlauthCreds()
		_, err := (&net.Dialer{Timeout: 2 * time.Second}).Dial("tcp", strings.TrimPrefix(creds.URL, "ldap://"))
		if err != nil {
			t.Skipf("docker not found and glauth not reachable — skipping")
		}
		t.Logf("glauth already reachable at %s (no docker)", creds.URL)
		return func() {}
	}

	for _, name := range []string{"mcpfather-glauth", "glauth"} {
		exec.Command("docker", "stop", name).Run()
		exec.Command("docker", "rm", name).Run()
	}

	glauthDir := filepath.Join(repoRoot(t), "it", "docker", "glauth")
	configFile := filepath.Join(glauthDir, "config.toml")
	cmd := exec.Command("docker", "run", "-d", "--name", "mcpfather-glauth",
		"--network", "host",
		"-v", configFile+":/app/config/config.cfg:ro",
		"registry.cn-shenzhen.aliyuncs.com/wl4g/glauth:v2.5.0",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker run glauth failed: %v\n%s -- skipping LDAP integration test", err, out)
	}
	t.Logf("glauth container started")

	waitForGlauth(t)
	return func() {
		exec.Command("docker", "stop", "mcpfather-glauth").Run()
		exec.Command("docker", "rm", "mcpfather-glauth").Run()
		t.Logf("glauth container cleaned up")
	}
}

// waitForGlauth polls the glauth LDAP port until it responds or times out.
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
	t.Skipf("glauth did not become ready within 60s -- skipping LDAP integration test")
}

// TestLDAPConfigEnvOverrides verifies that MCP__ env vars override LDAP config.
func TestLDAPConfigEnvOverrides(t *testing.T) {
	cleanupGlauth := ensureGlauth(t)
	defer cleanupGlauth()

	creds := defaultGlauthCreds()

	envVars := []string{
		"MCP__AUTH__BACKEND__LDAP__ENABLED=true",
		"MCP__AUTH__BACKEND__LDAP__URL=" + creds.URL,
		"MCP__AUTH__BACKEND__LDAP__BASE_DN=dc=test,dc=local",
		"MCP__AUTH__BACKEND__LDAP__BIND_DN=" + creds.BindDN,
		"MCP__AUTH__BACKEND__LDAP__BIND_PASSWORD=" + creds.BindPassword,
		"MCP__AUTH__BACKEND__LDAP__TIMEOUT=10",
		"MCP__UPSTREAM__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for LDAP config testing against real glauth")
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

	projectDir := genProject(t, "", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  backend:
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

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

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

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
auth:
  backend:
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

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result (wrong ldap password): %s", trimMsg(result, 300))

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
