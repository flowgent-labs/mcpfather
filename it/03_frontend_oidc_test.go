package tests

import (
	"bytes"
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

	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mcpHTTPCallWithAuth is like mcpHTTPCall but adds an Authorization header
// to every request (initialize, notification, and the actual tool call).
func mcpHTTPCallWithAuth(t *testing.T, baseURL, bearerToken, method string, params map[string]interface{}) (*http.Response, string) {
	t.Helper()

	authHeader := "Bearer " + bearerToken

	// Step 1: initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	body, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("initialize request failed: %v", err)
	}
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

	// If we got a 401, return early — don't try the notification.
	if resp.StatusCode == http.StatusUnauthorized {
		return resp, sessionID
	}

	// Step 2: send initialized notification
	if sessionID != "" {
		notifReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notifications/initialized",
		}
		body, _ = json.Marshal(notifReq)
		req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Mcp-Session-Id", sessionID)
		req.Header.Set("Authorization", authHeader)
		r, err := http.DefaultClient.Do(req)
		if err == nil {
			r.Body.Close()
		}
	}

	// Step 3: send the actual request
	mcpReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  method,
		"params":  params,
	}
	body, _ = json.Marshal(mcpReq)
	req, _ = http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	req.Header.Set("Authorization", authHeader)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("MCP %s request failed: %v", method, err)
	}
	return resp, sessionID
}

// makeHS256Token creates a JWT signed with HMAC-SHA256, for testing algorithm
// confusion prevention (server only accepts asymmetric algos from JWKS).
func makeHS256Token(t *testing.T, issuer, audience string) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": issuer,
		"sub": "test-user",
		"aud": audience,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("dummy-hmac-secret"))
	if err != nil {
		t.Fatalf("failed to sign HS256 token: %v", err)
	}
	return signed
}

// ---------------------------------------------------------------------------
// Test: 401 without token
// ---------------------------------------------------------------------------
func TestFrontend_401WithoutToken(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// Send a request without Authorization header
	resp, body := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d (body: %s)", resp.StatusCode, body)
	}
	wwwAuth := resp.Header.Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, `error="invalid_token"`) {
		t.Errorf("expected WWW-Authenticate with error=invalid_token, got: %s", wwwAuth)
	}
	if !strings.Contains(wwwAuth, "resource_metadata=") {
		t.Errorf("expected resource_metadata in WWW-Authenticate, got: %s", wwwAuth)
	}
	t.Logf("401 without token: %s", wwwAuth)
}

// ---------------------------------------------------------------------------
// Test: 401 with expired token
// ---------------------------------------------------------------------------
func TestFrontend_401ExpiredToken(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// Sign a token with exp in the past
	now := time.Now()
	expiredToken := oidc.SignToken(t, map[string]interface{}{
		"iss": oidc.Issuer(),
		"sub": "test-user",
		"aud": "mcpfather-frontend-test",
		"iat": now.Add(-2 * time.Hour).Unix(),
		"exp": now.Add(-1 * time.Hour).Unix(),
	})

	resp, _ := mcpHTTPCallWithAuth(t, baseURL, expiredToken, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", resp.StatusCode)
	}
	t.Logf("Expired token correctly rejected with 401")
}

// ---------------------------------------------------------------------------
// Test: 401 with wrong audience
// ---------------------------------------------------------------------------
func TestFrontend_401WrongAudience(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: correct-audience
`, mock.server.URL, oidc.Issuer())
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

	// Sign a token with wrong audience
	wrongAudToken := oidc.SignToken(t, map[string]interface{}{
		"iss": oidc.Issuer(),
		"sub": "test-user",
		"aud": "wrong-audience",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	resp, _ := mcpHTTPCallWithAuth(t, baseURL, wrongAudToken, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong audience, got %d", resp.StatusCode)
	}
	t.Logf("Wrong audience token correctly rejected with 401")
}

// ---------------------------------------------------------------------------
// Test: 401 with wrong issuer
// ---------------------------------------------------------------------------
func TestFrontend_401WrongIssuer(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// Sign a token with wrong issuer
	wrongIssToken := oidc.SignToken(t, map[string]interface{}{
		"iss": "https://evil-idp.example.com",
		"sub": "test-user",
		"aud": "mcpfather-frontend-test",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	resp, _ := mcpHTTPCallWithAuth(t, baseURL, wrongIssToken, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong issuer, got %d", resp.StatusCode)
	}
	t.Logf("Wrong issuer token correctly rejected with 401")
}

// ---------------------------------------------------------------------------
// Test: 401 with HS256 forgery (algorithm confusion attack)
// ---------------------------------------------------------------------------
func TestFrontend_401HS256Forgery(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// Create an HS256 token — the server only trusts RSA keys from JWKS,
	// so this must be rejected (algorithm confusion prevention, RFC 8725 §3.1).
	hs256Token := makeHS256Token(t, oidc.Issuer(), "mcpfather-frontend-test")

	resp, _ := mcpHTTPCallWithAuth(t, baseURL, hs256Token, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for HS256 forgery token, got %d", resp.StatusCode)
	}
	t.Logf("HS256 algorithm confusion token correctly rejected with 401")
}

// ---------------------------------------------------------------------------
// Test: 401 with malformed Authorization header (Basic auth)
// ---------------------------------------------------------------------------
func TestFrontend_401MalformedHeader(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// Send Basic auth instead of Bearer
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0") // base64("test:test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for Basic auth, got %d", resp.StatusCode)
	}
	t.Logf("Malformed (Basic) auth header correctly rejected with 401")
}

// ---------------------------------------------------------------------------
// Test: Device Code Flow — interactive AI coding agent (RFC 8628)
//
// Scenario for agents with human-in-the-loop (Claude Code, Copilot, etc.):
//  0. Start OIDC provider (testoidc, simulating Dex with device code support).
//  1. AI agent client calls MCP server without a token.
//  2. MCP server returns 401 with WWW-Authenticate header.
//  3. Client fetches RFC 9728 metadata → extracts authorization server URL.
//  4. Client discovers OIDC endpoints → finds device_authorization_endpoint.
//  5. Client POSTs to device/code endpoint → gets device_code + user_code.
//  6. User approves on another device (auto-approved in test provider).
//  7. Client polls token endpoint with device_code grant → receives JWT.
//  8. Client retries MCP call with JWT → MCP server validates via JWKS → success.
// ---------------------------------------------------------------------------
func TestFrontend_DeviceCodeFlow_InteractiveAgent(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather
`, mock.server.URL, oidc.Issuer())
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

	// --- Step 1-2: 401 + WWW-Authenticate ---
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("step 1: expected 401, got %d", resp.StatusCode)
	}
	wwwAuth := resp.Header.Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, "resource_metadata=") {
		t.Fatalf("step 2: missing resource_metadata in WWW-Authenticate: %s", wwwAuth)
	}
	t.Logf("Step 1-2: 401 received with WWW-Authenticate")

	// --- Step 3: RFC 9728 metadata ---
	metadataPath := extractResourceMetadata(t, wwwAuth)
	metadataURL := baseURL + metadataPath
	metadataResp, _ := http.Get(metadataURL)
	defer metadataResp.Body.Close()
	var metadata struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}
	json.NewDecoder(metadataResp.Body).Decode(&metadata)
	authServer := metadata.AuthorizationServers[0]
	t.Logf("Step 3: authorization server = %s", authServer)

	// --- Step 4: OIDC discovery — extract device_authorization_endpoint ---
	discURL := authServer + "/.well-known/openid-configuration"
	discResp, _ := http.Get(discURL)
	defer discResp.Body.Close()
	var discovery struct {
		Issuer                     string   `json:"issuer"`
		DeviceAuthEndpoint         string   `json:"device_authorization_endpoint"`
		TokenURL                   string   `json:"token_endpoint"`
		GrantTypes                 []string `json:"grant_types_supported"`
	}
	json.NewDecoder(discResp.Body).Decode(&discovery)
	if discovery.DeviceAuthEndpoint == "" {
		t.Fatal("step 4: OIDC discovery missing device_authorization_endpoint")
	}
	t.Logf("Step 4: discovery — issuer=%s device_endpoint=%s grant_types=%v",
		discovery.Issuer, discovery.DeviceAuthEndpoint, discovery.GrantTypes)

	// --- Step 5: Initiate device authorization ---
	deviceResp, _ := http.Post(discovery.DeviceAuthEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader("client_id=mcpfather-client&scope=openid"))
	defer deviceResp.Body.Close()
	if deviceResp.StatusCode != http.StatusOK {
		t.Fatalf("step 5: device/code returned %d", deviceResp.StatusCode)
	}
	var deviceCodeResp struct {
		DeviceCode string `json:"device_code"`
		UserCode   string `json:"user_code"`
		Interval   int    `json:"interval"`
	}
	json.NewDecoder(deviceResp.Body).Decode(&deviceCodeResp)
	t.Logf("Step 5: device code issued — user_code=%s interval=%d",
		deviceCodeResp.UserCode, deviceCodeResp.Interval)

	// --- Step 6: User approves (auto-approved by test provider) ---
	// In production the user visits verification_uri and enters the code.
	// The test OIDC provider auto-approves all device codes.
	t.Logf("Step 6: device auto-approved (test provider)")

	// --- Step 7: Poll token endpoint with device_code grant ---
	tokenBody := strings.NewReader(fmt.Sprintf(
		"grant_type=urn:ietf:params:oauth:grant-type:device_code&"+
			"device_code=%s&"+
			"client_id=mcpfather-client&"+
			"client_secret=mcpfather-secret",
		deviceCodeResp.DeviceCode))
	tokenResp, _ := http.Post(discovery.TokenURL, "application/x-www-form-urlencoded", tokenBody)
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(tokenResp.Body)
		t.Fatalf("step 7: token endpoint returned %d: %s", tokenResp.StatusCode, string(bodyBytes))
	}
	var tokenResult struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenResult)
	if tokenResult.AccessToken == "" {
		t.Fatal("step 7: token response missing access_token")
	}
	t.Logf("Step 7: JWT obtained via device_code grant (type=%s)", tokenResult.TokenType)

	// --- Step 8: Retry MCP call with JWT — success ---
	result := callNativeToolWithAuth(t, baseURL, tokenResult.AccessToken, "EchoHeaders", map[string]interface{}{})
	t.Logf("Step 8: MCP call result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("step 8: no request reached the mock upstream")
	}

	// Token Passthrough Prohibition
	upstreamAuth := mock.requests[0].Authorization
	if upstreamAuth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", upstreamAuth)
	}

	// Client token claims forwarded upstream (default: enable_client_token_claim_forward=true)
	sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub")
	if sub == "" {
		t.Error("expected X-MCP-Client-Token-Sub header, but it was missing or empty")
	} else {
		t.Logf("X-MCP-Client-Token-Sub forwarded: %s", sub)
	}

	t.Logf("Device code flow complete — frontend JWT validated, request forwarded upstream")
}

// ---------------------------------------------------------------------------
// Test: Client Credentials Flow — background AI agent (no human interaction)
//
// Scenario for fully-automated agents (OpenClaw, FlowGent, etc.):
//  0. Start OIDC provider (testoidc, simulating Dex with client_credentials).
//  1. AI agent client calls MCP server without a token.
//  2. MCP server returns 401 with WWW-Authenticate header.
//  3. Client fetches RFC 9728 metadata → extracts authorization server URL.
//  4. Client discovers OIDC endpoints → finds token_endpoint.
//  5. Client authenticates via client_credentials grant (machine-to-machine).
//  6. OIDC provider validates client_id/client_secret, signs JWT with RSA key.
//  7. Client retries MCP call with JWT → MCP server validates via JWKS → success.
// ---------------------------------------------------------------------------
func TestFrontend_ClientCredentialsFlow_BackgroundAgent(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather
`, mock.server.URL, oidc.Issuer())
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

	// --- Step 1-2: 401 + WWW-Authenticate ---
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("step 1: expected 401, got %d", resp.StatusCode)
	}
	wwwAuth := resp.Header.Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, "resource_metadata=") {
		t.Fatalf("step 2: missing resource_metadata in WWW-Authenticate: %s", wwwAuth)
	}
	t.Logf("Step 1-2: 401 received with WWW-Authenticate")

	// --- Step 3: RFC 9728 metadata ---
	metadataPath := extractResourceMetadata(t, wwwAuth)
	metadataURL := baseURL + metadataPath
	metadataResp, _ := http.Get(metadataURL)
	defer metadataResp.Body.Close()
	var metadata struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}
	json.NewDecoder(metadataResp.Body).Decode(&metadata)
	authServer := metadata.AuthorizationServers[0]
	t.Logf("Step 3: authorization server = %s", authServer)

	// --- Step 4: OIDC discovery ---
	discURL := authServer + "/.well-known/openid-configuration"
	discResp, _ := http.Get(discURL)
	defer discResp.Body.Close()
	var discovery struct {
		Issuer     string   `json:"issuer"`
		TokenURL   string   `json:"token_endpoint"`
		GrantTypes []string `json:"grant_types_supported"`
	}
	json.NewDecoder(discResp.Body).Decode(&discovery)
	t.Logf("Step 4: discovery — issuer=%s token_endpoint=%s grant_types=%v",
		discovery.Issuer, discovery.TokenURL, discovery.GrantTypes)

	// --- Step 5-6: client_credentials grant (machine-to-machine) ---
	tokenBody := strings.NewReader(
		"grant_type=client_credentials&" +
			"client_id=mcpfather-client&" +
			"client_secret=mcpfather-secret&" +
			"scope=openid")
	tokenResp, _ := http.Post(discovery.TokenURL, "application/x-www-form-urlencoded", tokenBody)
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(tokenResp.Body)
		t.Fatalf("step 5: token endpoint returned %d: %s", tokenResp.StatusCode, string(bodyBytes))
	}
	var tokenResult struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenResult)
	if tokenResult.AccessToken == "" {
		t.Fatal("step 5: token response missing access_token")
	}
	t.Logf("Step 5-6: JWT obtained via client_credentials grant (type=%s)", tokenResult.TokenType)

	// --- Step 7: Retry MCP call with JWT — success ---
	result := callNativeToolWithAuth(t, baseURL, tokenResult.AccessToken, "EchoHeaders", map[string]interface{}{})
	t.Logf("Step 7: MCP call result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("step 7: no request reached the mock upstream")
	}

	// Token Passthrough Prohibition
	upstreamAuth := mock.requests[0].Authorization
	if upstreamAuth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", upstreamAuth)
	}

	// Client token claims forwarded upstream (default: enable_client_token_claim_forward=true)
	sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub")
	if sub == "" {
		t.Error("expected X-MCP-Client-Token-Sub header, but it was missing or empty")
	} else {
		t.Logf("X-MCP-Client-Token-Sub forwarded: %s", sub)
	}

	t.Logf("Client credentials flow complete — frontend JWT validated, request forwarded upstream")
}

// ---------------------------------------------------------------------------
// Test: Client token claims forwarding (X-MCP-Client-Token-Sub / Email)
// ---------------------------------------------------------------------------
func TestFrontend_ClientTokenForwarding(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	// Default config: enable_client_token_claim_forward is true by default.
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather
`, mock.server.URL, oidc.Issuer())
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

	// Sign a token with both sub and email claims.
	token := oidc.SignToken(t, map[string]interface{}{
		"iss":   oidc.Issuer(),
		"sub":   "dev-ai-agent-42",
		"email": "ai-agent@enterprise.com",
		"aud":   "mcpfather",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	result := callNativeToolWithAuth(t, baseURL, token, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("no request reached the mock upstream")
	}

	sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub")
	if sub != "dev-ai-agent-42" {
		t.Errorf("expected X-MCP-Client-Token-Sub='dev-ai-agent-42', got '%s'", sub)
	}
	t.Logf("X-MCP-Client-Token-Sub: %s", sub)

	email := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Email")
	if email != "ai-agent@enterprise.com" {
		t.Errorf("expected X-MCP-Client-Token-Email='ai-agent@enterprise.com', got '%s'", email)
	}
	t.Logf("X-MCP-Client-Token-Email: %s", email)

	// Token Passthrough Prohibition still enforced
	if auth := mock.requests[0].Authorization; auth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", auth)
	}
}

// ---------------------------------------------------------------------------
// Test: Client token forwarding disabled via config
// ---------------------------------------------------------------------------
func TestFrontend_ClientTokenForwardingDisabled(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
  enable_client_token_claim_forward: false
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather
`, mock.server.URL, oidc.Issuer())
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

	token := oidc.SignToken(t, map[string]interface{}{
		"iss":   oidc.Issuer(),
		"sub":   "dev-ai-agent-42",
		"email": "ai-agent@enterprise.com",
		"aud":   "mcpfather",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})

	result := callNativeToolWithAuth(t, baseURL, token, "EchoHeaders", map[string]interface{}{})
	t.Logf("Tool result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("no request reached the mock upstream")
	}

	if sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub"); sub != "" {
		t.Errorf("X-MCP-Client-Token-Sub should NOT be forwarded when disabled, got '%s'", sub)
	}
	if email := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Email"); email != "" {
		t.Errorf("X-MCP-Client-Token-Email should NOT be forwarded when disabled, got '%s'", email)
	}
	t.Logf("Client token forwarding correctly disabled")
}

// extractResourceMetadata parses the resource_metadata URL from a
// WWW-Authenticate header value like:
//
//	Bearer error="invalid_token", resource_metadata="https://..."
func extractResourceMetadata(t *testing.T, wwwAuth string) string {
	t.Helper()
	// Find the resource_metadata="..." segment
	const prefix = `resource_metadata="`
	idx := strings.Index(wwwAuth, prefix)
	if idx < 0 {
		t.Fatalf("resource_metadata not found in WWW-Authenticate: %s", wwwAuth)
	}
	start := idx + len(prefix)
	// The value may be a relative path (not an absolute URL). Find the closing quote.
	end := strings.Index(wwwAuth[start:], `"`)
	if end < 0 {
		t.Fatalf("unterminated resource_metadata value: %s", wwwAuth)
	}
	return wwwAuth[start : start+end]
}

// callNativeToolWithAuth is like callNativeTool but uses mcpHTTPCallWithAuth.
func callNativeToolWithAuth(t *testing.T, baseURL, bearerToken, toolName string, args map[string]interface{}) string {
	t.Helper()
	resp, _ := mcpHTTPCallWithAuth(t, baseURL, bearerToken, "tools/call", map[string]interface{}{
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

// ---------------------------------------------------------------------------
// Test: Well-Known Protected Resource Metadata (RFC 9728 §3.1)
// ---------------------------------------------------------------------------
func TestFrontend_WellKnownMetadata(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
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

	// The metadata path is /.well-known/oauth-protected-resource for non-URL
	// resource identifiers (ProtectedResourceMetadataPath returns the base
	// well-known path when the resource is not an absolute URL).
	metadataURL := baseURL + "/.well-known/oauth-protected-resource"
	resp, err := http.Get(metadataURL)
	if err != nil {
		t.Fatalf("metadata request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for metadata endpoint, got %d", resp.StatusCode)
	}

	var metadata struct {
		Resource               string   `json:"resource"`
		AuthorizationServers   []string `json:"authorization_servers"`
		BearerMethodsSupported []string `json:"bearer_methods_supported"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}

	if metadata.Resource != "mcpfather-frontend-test" {
		t.Errorf("expected resource 'mcpfather-frontend-test', got %q", metadata.Resource)
	}
	if len(metadata.AuthorizationServers) == 0 || metadata.AuthorizationServers[0] != oidc.Issuer() {
		t.Errorf("expected authorization_servers to contain %q, got %v", oidc.Issuer(), metadata.AuthorizationServers)
	}
	found := false
	for _, m := range metadata.BearerMethodsSupported {
		if strings.EqualFold(m, "header") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected bearer_methods_supported to contain 'header', got %v", metadata.BearerMethodsSupported)
	}
	t.Logf("Well-known metadata: resource=%s auth_servers=%v methods=%v",
		metadata.Resource, metadata.AuthorizationServers, metadata.BearerMethodsSupported)
}

// ---------------------------------------------------------------------------
// Test: stdio mode ignores frontend auth (warning but no crash)
// ---------------------------------------------------------------------------
func TestFrontend_StdioNoAuth(t *testing.T) {
	oidc := startMockDexForDeprecatedGrants(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	binaryName := filepath.Base(projectDir)

	homeDir := t.TempDir()
	configYAML := fmt.Sprintf(`
upstream:
  endpoint: %s
auth:
  frontend:
    oidc:
      enabled: true
      issuer: %s
      audience: mcpfather-frontend-test
`, mock.server.URL, oidc.Issuer())
	writeCoreVirtualConfig(t, homeDir, binaryName, configYAML)

	cmd := exec.Command(binPath, "--transport", "stdio")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Pipe a valid initialize JSON-RPC message to stdin so ServeStdio doesn't
	// block forever, then close stdin to make it exit cleanly.
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start stdio server: %v", err)
	}

	// Send initialize and wait briefly for processing
	initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`
	stdinPipe.Write([]byte(initMsg + "\n"))
	time.Sleep(500 * time.Millisecond)

	// Close stdin to trigger ServeStdio shutdown
	stdinPipe.Close()
	cmd.Wait()

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "frontend") {
		t.Errorf("expected stderr warning about frontend auth ignored in stdio, got: %s", stderr)
	}
	t.Logf("Stdio warning: %s", strings.TrimSpace(stderr))
}
