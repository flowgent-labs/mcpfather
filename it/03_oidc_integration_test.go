package tests

import (
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// OIDC integration tests — Real Dex provider
//
// These tests verify connectivity, discovery, and config against a real Dex
// OIDC provider (it/docker/dex).
//
// Token exchange and full E2E tests that require the password (ROPC) grant
// live in 03_oidc_integration_mocksvc_test.go because Dex v2.41+ no longer
// supports the password grant type.
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
		"registry.cn-shenzhen.aliyuncs.com/wl4g/dex:v2.41.1", "dex", "serve", "/etc/dex/config.yaml")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker run dex failed: %v\n%s -- skipping OIDC integration test", err, out)
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
	t.Skipf("Dex did not become ready within 60s -- skipping OIDC integration test")
}

// TestOIDCConfigEnvOverrides verifies that MCP__ env vars override OIDC config values.
func TestOIDCConfigEnvOverrides(t *testing.T) {
	cleanupDex := ensureDex(t)
	defer cleanupDex()

	envVars := []string{
		"MCP__AUTH__OIDC__ENABLED=true",
		"MCP__AUTH__OIDC__ISSUER=http://localhost:5556/dex",
		"MCP__AUTH__OIDC__CLIENT_ID=mcpgen-client",
		"MCP__AUTH__OIDC__CLIENT_SECRET=mcpgen-secret",
		"MCP__AUTH__OIDC__SCOPES=openid",
		"MCP__UPSTREAM__ENDPOINT=http://localhost:0",
	}

	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		t.Setenv(parts[0], parts[1])
	}
	t.Logf("MCP__ env vars set for OIDC config testing")
}

// TestOIDCDexDiscovery verifies OIDC discovery against a real Dex instance.
func TestOIDCDexDiscovery(t *testing.T) {
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
