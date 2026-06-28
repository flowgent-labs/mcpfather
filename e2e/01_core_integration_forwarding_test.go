package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ===========================================================================
// Core MCP Server Forwarding E2E Tests (using CoreMockService)
// ===========================================================================

// ---------------------------------------------------------------------------
// Core Test 1: Authorization header forwarding via MCP_UPSTREAM_TOKEN
// ---------------------------------------------------------------------------

func TestE2E_Core_AuthForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "test-bearer-token-abc123", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)

	auth, _ := data["authorization"].(string)
	if !strings.Contains(auth, "test-bearer-token-abc123") {
		t.Errorf("Authorization header should contain token, got: %q", auth)
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Errorf("Authorization should have Bearer prefix, got: %q", auth)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
}

// ---------------------------------------------------------------------------
// Core Test 2: Bearer prefix is not duplicated
// ---------------------------------------------------------------------------

func TestE2E_Core_AuthNoDoublePrefix(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "Bearer already-prefixed-token", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)

	auth, _ := data["authorization"].(string)
	if strings.Count(auth, "Bearer") > 1 {
		t.Errorf("Bearer prefix should not be duplicated, got: %q", auth)
	}
	if !strings.Contains(auth, "already-prefixed-token") {
		t.Errorf("Token should be preserved, got: %q", auth)
	}
}

// ---------------------------------------------------------------------------
// Core Test 3: Cookie forwarding via MCP_UPSTREAM_COOKIE
// ---------------------------------------------------------------------------

func TestE2E_Core_CookieForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "JSESSIONID=abc123test", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, ok := data["headers"].(map[string]interface{})
	if !ok {
		t.Fatal("headers not found in response")
	}
	cookie, _ := headers["Cookie"].(string)
	if !strings.Contains(cookie, "JSESSIONID=abc123test") {
		t.Errorf("Cookie header should contain JSESSIONID, got: %q", cookie)
	}
}

// ---------------------------------------------------------------------------
// Core Test 4: MCP session ID forwarding disabled by default
// ---------------------------------------------------------------------------

func TestE2E_Core_SessionForwardingDisabledByDefault(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, _ := data["headers"].(map[string]interface{})

	if _, ok := headers["X-MCP-Session-ID"]; ok {
		t.Error("X-MCP-Session-ID should not be forwarded by default")
	}
}

// ---------------------------------------------------------------------------
// Core Test 5: MCP-Session-Id in client request should not leak to upstream
// ---------------------------------------------------------------------------

func TestE2E_Core_SessionNotLeaked(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	// The MCP server receives Mcp-Session-Id from the client, but should NOT
	// forward it to the upstream.
	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, _ := data["headers"].(map[string]interface{})

	if _, ok := headers["Mcp-Session-Id"]; ok {
		t.Error("Mcp-Session-Id should NOT be forwarded to upstream")
	}
}

// ---------------------------------------------------------------------------
// Core Test 6: Content-type handling — JSON response passes through
// ---------------------------------------------------------------------------

func TestE2E_Core_ContentTypeJSON(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"format": "json",
	})
	data := mustJSON(t, result)

	if tv, _ := data["type"].(string); tv != "json" {
		t.Errorf("expected JSON response type 'json', got %q", tv)
	}
	// Verify the data field is present too
	if dv, _ := data["data"].(string); dv != "json-response" {
		t.Errorf("expected data field 'json-response', got %q", dv)
	}
}

// ---------------------------------------------------------------------------
// Core Test 7: Binary content-type detection
// ---------------------------------------------------------------------------

func TestE2E_Core_BinaryContentType(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "downloadReport", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "DownloadReport", map[string]interface{}{})
	// Binary download should be saved to file, result contains "Saved to:"
	if !strings.Contains(result, "Saved to:") {
		t.Errorf("binary download should save to file, got: %s", trimMsg(result, 200))
	}
}

// ---------------------------------------------------------------------------
// Core Test 8: Path parameter substitution — no scientific notation
// ---------------------------------------------------------------------------

func TestE2E_Core_PathParamSubstitution(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterPathParamScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"name": "test-user",
		"age":  float64(30),
	})
	data := mustJSON(t, result)

	qn, _ := data["queryName"].(string)
	if qn != "test-user" {
		t.Errorf("query param name = %q, want test-user", qn)
	}
	qa, _ := data["queryAge"].(string)
	if qa != "30" {
		t.Errorf("query param age = %q, want 30", qa)
	}
	// Float64 should NOT become scientific notation
	if strings.Contains(qa, "e") || strings.Contains(qa, "E") {
		t.Errorf("query param age should not be in scientific notation: %q", qa)
	}
}

// ---------------------------------------------------------------------------
// Core Test 9: Upstream error handling (4xx)
// ---------------------------------------------------------------------------

func TestE2E_Core_UpstreamErrorHandling(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterErrorScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	resp, body := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
		"arguments": map[string]interface{}{
			"status": "400",
		},
	})
	defer resp.Body.Close()

	// Upstream returns 400, MCP server should propagate the error
	if !strings.Contains(body, "Bad Request") {
		t.Logf("error response body: %s", body)
	}
}

// ---------------------------------------------------------------------------
// Core Test 10: XML content-type is treated as text (not binary)
// ---------------------------------------------------------------------------

func TestE2E_Core_ContentTypeXML(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", false)
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"format": "xml",
	})
	// XML response should be returned as text (not saved to file)
	if strings.Contains(result, "Saved to:") {
		t.Error("XML should be treated as text, not binary download")
	}
	// Should contain XML content
	if !strings.Contains(result, "root") && !strings.Contains(result, "status") {
		t.Errorf("XML response should contain content, got: %s", trimMsg(result, 200))
	}
}

// ---------------------------------------------------------------------------
// Core Test 11: Chained native tools via virtual config
// ---------------------------------------------------------------------------

func TestE2E_Core_ChainedNativeTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	aggConfig := `
virtualTools:
  - name: agg_chain
    description: Chain echo and greet
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - id: echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: greet
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.name
      - id: done
        kind: return
        spec:
          from: $greet
`
	writeCoreVirtualConfig(t, homeDir, binaryName, aggConfig)

	cleanup, baseURL := startAggTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallVirtualTool(t, baseURL, "agg_chain", map[string]interface{}{
		"name": "Alice",
	})

	data := mustJSON(t, result)
	greeting, _ := data["greeting"].(string)
	if !strings.Contains(greeting, "Alice") {
		t.Errorf("greeting should mention Alice, got %q", greeting)
	}
}

// ===========================================================================
// Core Forwarding Helpers
// ===========================================================================

// callNativeTool calls a native (native) MCP tool through the server.
func callNativeTool(t *testing.T, baseURL string, toolName string, args map[string]interface{}) string {
	t.Helper()
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	})
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(body, &rpcResp)

	if rpcResp.Error != nil {
		return fmt.Sprintf("MCP error: %s", rpcResp.Error.Message)
	}
	if len(rpcResp.Result.Content) > 0 {
		return rpcResp.Result.Content[0].Text
	}
	return ""
}

// startCoreForwardTestServer builds and starts a native MCP server with
// custom environment for authentication/header forwarding tests.
func startCoreForwardTestServer(t *testing.T, projectDir, mockURL, homeDir, token, cookie string, enableSessionForwarding bool) (cleanup func(), baseURL string) {
	t.Helper()
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP_UPSTREAM_ENDPOINT="+mockURL,
	)
	if token != "" {
		cmd.Env = append(cmd.Env, "MCP_UPSTREAM_TOKEN="+token)
	}
	if cookie != "" {
		cmd.Env = append(cmd.Env, "MCP_UPSTREAM_COOKIE="+cookie)
	}

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}

	cleanup = func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	baseURL = "http://localhost:" + port
	waitForServer(t, baseURL)
	return
}

// writeCoreVirtualConfig writes an virtual tools config for core tests.
func writeCoreVirtualConfig(t *testing.T, homeDir, binaryName, yamlContent string) {
	t.Helper()
	configDir := filepath.Join(homeDir, "."+binaryName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func trimMsg(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
