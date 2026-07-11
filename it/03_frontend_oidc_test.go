package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ===========================================================================
// Shared helpers
// ===========================================================================

func mcpHTTPCallWithAuth(t *testing.T, baseURL, bearerToken, method string, params map[string]interface{}) (*http.Response, string) {
	t.Helper()

	authHeader := "Bearer " + bearerToken

	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test-client", "version": "1.0.0"},
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

	if resp.StatusCode == http.StatusUnauthorized {
		return resp, sessionID
	}

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

func extractResourceMetadata(t *testing.T, wwwAuth string) string {
	t.Helper()
	const prefix = `resource_metadata="`
	idx := strings.Index(wwwAuth, prefix)
	if idx < 0 {
		t.Fatalf("resource_metadata not found in WWW-Authenticate: %s", wwwAuth)
	}
	start := idx + len(prefix)
	end := strings.Index(wwwAuth[start:], `"`)
	if end < 0 {
		t.Fatalf("unterminated resource_metadata value: %s", wwwAuth)
	}
	return wwwAuth[start : start+end]
}

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

// ===========================================================================
// Keycloak OIDC provider infrastructure
//
// Keycloak 26.7.0 supports client_credentials and device_code (RFC 8628).
// Cookie handling: Keycloak sets the Secure flag on cookies even in dev mode.
// Since tests use http://, Go's cookiejar won't send Secure cookies.
// We work around this with manual cookie forwarding.
// ===========================================================================

func ensureKeycloak(t *testing.T) (issuer string, cleanup func()) {
	t.Helper()

	if !dockerAvailable() {
		resp, err := http.Get("http://127.0.0.1:8080/realms/master/.well-known/openid-configuration")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Skipf("docker not found and Keycloak not reachable -- skipping")
		}
		if resp != nil {
			resp.Body.Close()
		}
		issuer = "http://127.0.0.1:8080/realms/master"
		t.Logf("Keycloak already reachable (no docker)")
		return issuer, func() {}
	}

	for _, name := range []string{"mcpfather-keycloak"} {
		exec.Command("docker", "stop", name).Run()
		exec.Command("docker", "rm", name).Run()
	}

	keycloakDir := filepath.Join(repoRoot(t), "it", "docker", "keycloak")
	importFile := filepath.Join(keycloakDir, "realm.json")

	cmd := exec.Command("docker", "run", "-d", "--name", "mcpfather-keycloak",
		"--network", "host",
		"--hostname", "127.0.0.1",
		"-e", "KC_BOOTSTRAP_ADMIN_USERNAME=admin",
		"-e", "KC_BOOTSTRAP_ADMIN_PASSWORD=admin",
		"registry.cn-shenzhen.aliyuncs.com/wl4g/keycloak:26.7.0",
		"start-dev", "--hostname=127.0.0.1", "--hostname-strict=false",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker run keycloak failed: %v\n%s -- skipping Keycloak integration test", err, out)
	}
	t.Logf("Keycloak container started")

	cleanup = func() {
		exec.Command("docker", "stop", "mcpfather-keycloak").Run()
		exec.Command("docker", "rm", "mcpfather-keycloak").Run()
	}

	issuer = "http://127.0.0.1:8080/realms/master"

	if !waitForURL(t, issuer+"/.well-known/openid-configuration", 120) {
		cleanup()
		t.Skipf("Keycloak did not become ready within 120s -- skipping")
	}

	adminToken := keycloakAdminToken(t)
	setupKeycloakTestRealm(t, adminToken, importFile)

	t.Logf("Keycloak ready at %s", issuer)
	return issuer, cleanup
}

func waitForURL(t *testing.T, u string, maxSec int) bool {
	t.Helper()
	for i := 0; i < maxSec*2; i++ {
		resp, err := http.Get(u)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func keycloakAdminToken(t *testing.T) string {
	t.Helper()
	resp, err := http.PostForm("http://127.0.0.1:8080/realms/master/protocol/openid-connect/token",
		url.Values{
			"grant_type": {"password"},
			"client_id":  {"admin-cli"},
			"username":   {"admin"},
			"password":   {"admin"},
		})
	if err != nil {
		t.Fatalf("admin token request: %v", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.AccessToken == "" {
		t.Fatal("admin token response missing access_token")
	}
	return result.AccessToken
}

func setupKeycloakTestRealm(t *testing.T, adminToken, importFile string) {
	t.Helper()
	authHdr := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
	}

	if importFile != "" {
		if _, err := os.Stat(importFile); err == nil {
			realmJSON, err := os.ReadFile(importFile)
			if err == nil {
				req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/admin/realms", bytes.NewReader(realmJSON))
				authHdr(req)
				resp, err := http.DefaultClient.Do(req)
				if err == nil {
					resp.Body.Close()
					if resp.StatusCode == http.StatusCreated {
						t.Logf("Keycloak realm imported from %s", importFile)
						return
					}
				}
			}
		}
	}

	doJSON := func(method, urlPath string, body []byte) int {
		req, _ := http.NewRequest(method, "http://127.0.0.1:8080/"+urlPath, bytes.NewReader(body))
		authHdr(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Logf("Admin API %s %s: %v", method, urlPath, err)
			return 0
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	// Create client
	clientJSON := `{
		"clientId": "mcpfather-client",
		"secret": "mcpfather-secret",
		"standardFlowEnabled": false,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"publicClient": false,
		"attributes": {
			"oauth2.device.authorization.grant.enabled": "true"
		}
	}`
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/admin/realms/master/clients?clientId=mcpfather-client", nil)
	authHdr(req)
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		var clients []struct{ ID string }
		json.NewDecoder(resp.Body).Decode(&clients)
		resp.Body.Close()
		if len(clients) > 0 {
			t.Logf("Keycloak client mcpfather-client already exists")
		} else {
			doJSON("POST", "admin/realms/master/clients", []byte(clientJSON))
		}
	}

	// Create user
	userJSON := `{
		"username": "testuser",
		"email": "testuser@example.com",
		"enabled": true,
		"emailVerified": true,
		"credentials": [{"type": "password", "value": "testpass", "temporary": false}]
	}`
	req, _ = http.NewRequest("GET", "http://127.0.0.1:8080/admin/realms/master/users?username=testuser", nil)
	authHdr(req)
	resp, err = http.DefaultClient.Do(req)
	if err == nil {
		var users []struct{ ID string }
		json.NewDecoder(resp.Body).Decode(&users)
		resp.Body.Close()
		if len(users) > 0 {
			t.Logf("Keycloak user testuser already exists")
		} else {
			doJSON("POST", "admin/realms/master/users", []byte(userJSON))
		}
	}

	// Add audience mapper so tokens include "mcpfather" in aud
	req, _ = http.NewRequest("GET", "http://127.0.0.1:8080/admin/realms/master/clients?clientId=mcpfather-client", nil)
	authHdr(req)
	resp, err = http.DefaultClient.Do(req)
	if err == nil {
		var clients []struct{ ID string }
		json.NewDecoder(resp.Body).Decode(&clients)
		resp.Body.Close()
		if len(clients) > 0 {
			clientUUID := clients[0].ID
			req, _ = http.NewRequest("GET",
				"http://127.0.0.1:8080/admin/realms/master/clients/"+clientUUID+"/protocol-mappers/models", nil)
			authHdr(req)
			resp, err = http.DefaultClient.Do(req)
			if err == nil {
				var mappers []struct{ Name string }
				json.NewDecoder(resp.Body).Decode(&mappers)
				resp.Body.Close()
				hasAudMapper := false
				for _, m := range mappers {
					if m.Name == "mcpfather-audience" {
						hasAudMapper = true
						break
					}
				}
				if !hasAudMapper {
					mapperJSON := `{"name":"mcpfather-audience","protocol":"openid-connect","protocolMapper":"oidc-audience-mapper","config":{"included.client.audience":"mcpfather","id.token.claim":"false","access.token.claim":"true"}}`
					doJSON("POST", "admin/realms/master/clients/"+clientUUID+"/protocol-mappers/models", []byte(mapperJSON))
				}
			}
		}
	}
}

func keycloakClientCredentialsToken(t *testing.T, issuer string) string {
	t.Helper()
	resp, err := http.PostForm(issuer+"/protocol/openid-connect/token",
		url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"mcpfather-client"},
			"client_secret": {"mcpfather-secret"},
			"scope":         {"openid"},
		})
	if err != nil {
		t.Fatalf("client_credentials request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("client_credentials token: HTTP %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.AccessToken == "" {
		t.Fatal("client_credentials response missing access_token")
	}
	return result.AccessToken
}

// ---------------------------------------------------------------------------
// Cookie helpers — work around Keycloak's Secure flag on dev-mode cookies
// ---------------------------------------------------------------------------

type cookieJar struct {
	cookies map[string]string
	client  *http.Client
}

func newCookieJar() *cookieJar {
	return &cookieJar{
		cookies: make(map[string]string),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (j *cookieJar) extractCookies(resp *http.Response) *cookieJar {
	for _, sc := range resp.Header["Set-Cookie"] {
		parts := strings.SplitN(sc, ";", 2)
		if len(parts) > 0 {
			kv := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
			if len(kv) == 2 {
				j.cookies[kv[0]] = kv[1]
			}
		}
	}
	return j
}

func (j *cookieJar) header() string {
	if len(j.cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(j.cookies))
	for name, val := range j.cookies {
		parts = append(parts, name+"="+val)
	}
	return strings.Join(parts, "; ")
}

func (j *cookieJar) addCookies(req *http.Request) {
	if h := j.header(); h != "" {
		req.Header.Set("Cookie", h)
	}
	req.Header.Set("User-Agent", "mcpfather-test/1.0")
}

func (j *cookieJar) doGet(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, nil)
	j.addCookies(req)
	resp, err := j.client.Do(req)
	if err == nil {
		j.extractCookies(resp)
	}
	return resp, err
}

func (j *cookieJar) doPostForm(u string, data url.Values) (*http.Response, error) {
	req, _ := http.NewRequest("POST", u, strings.NewReader(data.Encode()))
	j.addCookies(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := j.client.Do(req)
	if err == nil {
		j.extractCookies(resp)
	}
	return resp, err
}

// ---------------------------------------------------------------------------
// Keycloak device flow automation
// ---------------------------------------------------------------------------

func keycloakApproveDeviceCode(t *testing.T, issuer, userCode string) {
	t.Helper()
	base := "http://127.0.0.1:8080"

	j := newCookieJar()

	verificationURI := fmt.Sprintf("%s/realms/master/device?user_code=%s", base, userCode)
	resp, err := j.doGet(verificationURI)
	if err != nil {
		t.Fatalf("get verification_uri_complete: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Fatalf("step 1: expected redirect from verification_uri_complete, got %d", resp.StatusCode)
	}

	re := regexp.MustCompile(`<form\b[^>]*\saction="([^"]*)"`)
	for range 10 {
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			if loc == "" {
				t.Fatalf("redirect status %d but no Location header", resp.StatusCode)
			}
			if strings.HasPrefix(loc, "/") {
				loc = base + loc
			}
			resp.Body.Close()
			resp, err = j.doGet(loc)
			if err != nil {
				t.Fatalf("follow redirect %s: %v", loc, err)
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			resp.Body.Close()
			t.Logf("Non-2xx response: status=%d", resp.StatusCode)
			break
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		match := re.FindSubmatch(body)
		if match == nil {
			break
		}
		action := strings.ReplaceAll(string(match[1]), "&amp;", "&")
		if !strings.HasPrefix(action, "http") {
			action = base + action
		}

		if strings.Contains(action, "consent") || strings.Contains(string(body), "OAUTH_GRANT") {
			resp, err = j.doPostForm(action, url.Values{"accept": {"Yes"}})
		} else {
			resp, err = j.doPostForm(action, url.Values{
				"username":     {"testuser"},
				"password":     {"testpass"},
				"credentialId": {""},
			})
		}
		if err != nil {
			t.Fatalf("post form %s: %v", action, err)
		}
		resp.Body.Close()
	}

	if _, ok := j.cookies["KEYCLOAK_SESSION"]; !ok {
		t.Fatalf("device approval did not produce KEYCLOAK_SESSION cookie; cookies=%v", j.cookies)
	}
	t.Logf("Device code %s approved (session established)", userCode)
}

func keycloakPollDeviceToken(t *testing.T, issuer, deviceCode string) string {
	t.Helper()
	for attempt := 0; attempt < 30; attempt++ {
		resp, err := http.PostForm(issuer+"/protocol/openid-connect/token",
			url.Values{
				"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code":   {deviceCode},
				"client_id":     {"mcpfather-client"},
				"client_secret": {"mcpfather-secret"},
			})
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			var result struct {
				AccessToken string `json:"access_token"`
			}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()
			if result.AccessToken != "" {
				t.Logf("Device token obtained after %d polls", attempt+1)
				return result.AccessToken
			}
			t.Fatal("token response missing access_token")
		}
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		resp.Body.Close()
		if errResp.Error == "slow_down" {
			time.Sleep(5 * time.Second)
		} else if errResp.Error == "authorization_pending" {
			time.Sleep(2 * time.Second)
		} else {
			t.Fatalf("device token error: %s", errResp.Error)
		}
	}
	t.Fatal("device token polling timed out")
	return ""
}

func keycloakInitDeviceAuth(t *testing.T, issuer string) (deviceCode, userCode string) {
	t.Helper()
	resp, err := http.PostForm(issuer+"/protocol/openid-connect/auth/device",
		url.Values{
			"client_id":     {"mcpfather-client"},
			"client_secret": {"mcpfather-secret"},
			"scope":         {"openid"},
		})
	if err != nil {
		t.Fatalf("device auth: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("device auth: HTTP %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		DeviceCode string `json:"device_code"`
		UserCode   string `json:"user_code"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.DeviceCode, result.UserCode
}

// ===========================================================================
// Real Keycloak tests — auth rejection, metadata, RFC 8628, M2M
// ===========================================================================

// startMCPServer is a helper that builds, configures, and starts a generated
// MCP server process in HTTP mode, returning the base URL and a cleanup func.
func startMCPServer(t *testing.T, issuer, audience string) (baseURL string, cleanup func(), mock *mockUpstream) {
	t.Helper()

	mock = startMockUpstream(okHandler())

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
      audience: %s
`, mock.server.URL, issuer, audience)
	writeCoreVirtualConfig(t, homeDir, binaryName, configYAML)

	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))
	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}

	baseURL = "http://localhost:" + port
	waitForServer(t, baseURL)

	cleanup = func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
		mock.Close()
	}
	return baseURL, cleanup, mock
}

// TestFrontend_401WithoutToken verifies the MCP server returns 401 when no
// Authorization header is present.
func TestFrontend_401WithoutToken(t *testing.T) {
	issuer, cleanupKC := ensureKeycloak(t)
	defer cleanupKC()

	baseURL, cleanupSrv, _ := startMCPServer(t, issuer, "mcpfather")
	defer cleanupSrv()

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

// TestFrontend_401MalformedHeader verifies the MCP server rejects non-Bearer
// Authorization headers (e.g. Basic auth) with 401.
func TestFrontend_401MalformedHeader(t *testing.T) {
	issuer, cleanupKC := ensureKeycloak(t)
	defer cleanupKC()

	baseURL, cleanupSrv, _ := startMCPServer(t, issuer, "mcpfather")
	defer cleanupSrv()

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
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
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

// TestFrontend_401HS256Forgery verifies algorithm confusion prevention:
// the server only trusts RSA keys from JWKS and must reject HMAC-signed tokens.
func TestFrontend_401HS256Forgery(t *testing.T) {
	issuer, cleanupKC := ensureKeycloak(t)
	defer cleanupKC()

	baseURL, cleanupSrv, _ := startMCPServer(t, issuer, "mcpfather")
	defer cleanupSrv()

	// Create an HS256 token with Keycloak's issuer — the server must
	// reject it because Keycloak's JWKS only contains RSA keys.
	hs256Token := makeHS256Token(t, issuer, "mcpfather")

	resp, _ := mcpHTTPCallWithAuth(t, baseURL, hs256Token, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for HS256 forgery token, got %d", resp.StatusCode)
	}
	t.Logf("HS256 algorithm confusion token correctly rejected with 401")
}

// TestFrontend_WellKnownMetadata verifies the RFC 9728 protected resource
// metadata endpoint.
func TestFrontend_WellKnownMetadata(t *testing.T) {
	issuer, cleanupKC := ensureKeycloak(t)
	defer cleanupKC()

	baseURL, cleanupSrv, _ := startMCPServer(t, issuer, "mcpfather-frontend-test")
	defer cleanupSrv()

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
	if len(metadata.AuthorizationServers) == 0 || metadata.AuthorizationServers[0] != issuer {
		t.Errorf("expected authorization_servers to contain %q, got %v", issuer, metadata.AuthorizationServers)
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

// TestFrontend_StdioNoAuth verifies that stdio transport emits a warning
// about frontend auth being ignored (rather than crashing).
func TestFrontend_StdioNoAuth(t *testing.T) {
	issuer, cleanupKC := ensureKeycloak(t)
	defer cleanupKC()

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
`, mock.server.URL, issuer)
	writeCoreVirtualConfig(t, homeDir, binaryName, configYAML)

	cmd := exec.Command(binPath, "--transport", "stdio")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start stdio server: %v", err)
	}

	initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`
	stdinPipe.Write([]byte(initMsg + "\n"))
	time.Sleep(500 * time.Millisecond)

	stdinPipe.Close()
	cmd.Wait()

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "frontend") {
		t.Errorf("expected stderr warning about frontend auth ignored in stdio, got: %s", stderr)
	}
	t.Logf("Stdio warning: %s", strings.TrimSpace(stderr))
}

// TestFrontend_DeviceCodeFlow_Keycloak tests the full RFC 8628 device_code
// flow against a real Keycloak container.
func TestFrontend_DeviceCodeFlow_Keycloak(t *testing.T) {
	issuer, cleanup := ensureKeycloak(t)
	defer cleanup()

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
`, mock.server.URL, issuer)
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

	// Step 1-2: 401 + WWW-Authenticate
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

	// Step 3: RFC 9728 metadata
	metadataPath := extractResourceMetadata(t, wwwAuth)
	metadataURL := baseURL + metadataPath
	resp, err := http.Get(metadataURL)
	if err != nil {
		t.Fatalf("metadata request: %v", err)
	}
	defer resp.Body.Close()
	var metadata struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}
	json.NewDecoder(resp.Body).Decode(&metadata)
	authServer := metadata.AuthorizationServers[0]
	t.Logf("Step 3: authorization server = %s", authServer)

	// Step 4: OIDC discovery
	discURL := authServer + "/.well-known/openid-configuration"
	resp, err = http.Get(discURL)
	if err != nil {
		t.Fatalf("discovery request: %v", err)
	}
	defer resp.Body.Close()
	var discovery struct {
		Issuer             string `json:"issuer"`
		DeviceAuthEndpoint string `json:"device_authorization_endpoint"`
		TokenURL           string `json:"token_endpoint"`
	}
	json.NewDecoder(resp.Body).Decode(&discovery)
	if discovery.DeviceAuthEndpoint == "" {
		t.Fatal("step 4: OIDC discovery missing device_authorization_endpoint")
	}
	t.Logf("Step 4: discovery — issuer=%s device_endpoint=%s",
		discovery.Issuer, discovery.DeviceAuthEndpoint)

	// Step 5: Initiate device authorization on Keycloak
	deviceCode, userCode := keycloakInitDeviceAuth(t, issuer)
	t.Logf("Step 5: device code issued — user_code=%s", userCode)

	// Step 6: User approves on device page (automated)
	keycloakApproveDeviceCode(t, issuer, userCode)
	t.Logf("Step 6: device approved by user")

	// Step 7: Poll token endpoint
	token := keycloakPollDeviceToken(t, issuer, deviceCode)
	t.Logf("Step 7: JWT obtained via device_code grant")

	// Step 8: Retry MCP call with JWT — success
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3-part JWT, got %d parts", len(parts))
	}

	result := callNativeToolWithAuth(t, baseURL, token, "EchoHeaders", map[string]interface{}{})
	t.Logf("Step 8: MCP call result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("step 8: no request reached the mock upstream")
	}

	// Token Passthrough Prohibition
	upstreamAuth := mock.requests[0].Authorization
	if upstreamAuth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", upstreamAuth)
	}

	sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub")
	if sub == "" {
		t.Error("expected X-MCP-Client-Token-Sub header, but it was missing or empty")
	} else {
		t.Logf("X-MCP-Client-Token-Sub forwarded: %s", sub)
	}

	t.Logf("Device code flow complete — real Keycloak JWT validated, request forwarded upstream")
}

// TestFrontend_ClientCredentialsFlow_Keycloak tests the full client_credentials
// (machine-to-machine) flow against a real Keycloak container.
func TestFrontend_ClientCredentialsFlow_Keycloak(t *testing.T) {
	issuer, cleanup := ensureKeycloak(t)
	defer cleanup()

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
`, mock.server.URL, issuer)
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

	// Step 1-2: 401 + WWW-Authenticate
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

	// Step 3: RFC 9728 metadata
	metadataPath := extractResourceMetadata(t, wwwAuth)
	metadataURL := baseURL + metadataPath
	resp, err := http.Get(metadataURL)
	if err != nil {
		t.Fatalf("metadata request: %v", err)
	}
	defer resp.Body.Close()
	var metadata struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}
	json.NewDecoder(resp.Body).Decode(&metadata)
	authServer := metadata.AuthorizationServers[0]
	t.Logf("Step 3: authorization server = %s", authServer)

	// Step 4: OIDC discovery
	discURL := authServer + "/.well-known/openid-configuration"
	resp, err = http.Get(discURL)
	if err != nil {
		t.Fatalf("discovery request: %v", err)
	}
	defer resp.Body.Close()
	var discovery struct {
		Issuer   string `json:"issuer"`
		TokenURL string `json:"token_endpoint"`
	}
	json.NewDecoder(resp.Body).Decode(&discovery)
	t.Logf("Step 4: discovery — issuer=%s token_endpoint=%s",
		discovery.Issuer, discovery.TokenURL)

	// Step 5-6: client_credentials grant (machine-to-machine)
	token := keycloakClientCredentialsToken(t, issuer)
	t.Logf("Step 5-6: JWT obtained via client_credentials grant")

	// Step 7: Retry MCP call with JWT — success
	result := callNativeToolWithAuth(t, baseURL, token, "EchoHeaders", map[string]interface{}{})
	t.Logf("Step 7: MCP call result: %s", trimMsg(result, 300))

	if mock.requestCount() == 0 {
		t.Fatal("step 7: no request reached the mock upstream")
	}

	// Token Passthrough Prohibition
	upstreamAuth := mock.requests[0].Authorization
	if upstreamAuth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", upstreamAuth)
	}

	sub := mock.requests[0].Headers.Get("X-Mcp-Client-Token-Sub")
	if sub == "" {
		t.Error("expected X-MCP-Client-Token-Sub header, but it was missing or empty")
	} else {
		t.Logf("X-MCP-Client-Token-Sub forwarded: %s", sub)
	}

	t.Logf("Client credentials flow complete — real Keycloak JWT validated, request forwarded upstream")
}

// ===========================================================================
// Mock OIDC provider tests — negative cases Keycloak can't produce
//
// These tests use startMockOIDCServer (from 03_backend_oidc_test.go) for
// custom JWT crafting via /sign: expired tokens, wrong audience, wrong
// issuer, and custom claim values for token forwarding verification.
// Keycloak cannot produce these scenarios.
// ===========================================================================

// TestFrontend_401ExpiredToken verifies 401 rejection for an expired JWT.
func TestFrontend_401ExpiredToken(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	baseURL, cleanupSrv, _ := startMCPServer(t, oidc.Issuer(), "mcpfather-frontend-test")
	defer cleanupSrv()

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

// TestFrontend_401WrongAudience verifies 401 rejection when audience doesn't match.
func TestFrontend_401WrongAudience(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	baseURL, cleanupSrv, _ := startMCPServer(t, oidc.Issuer(), "correct-audience")
	defer cleanupSrv()

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

// TestFrontend_401WrongIssuer verifies 401 rejection when issuer doesn't match.
func TestFrontend_401WrongIssuer(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	baseURL, cleanupSrv, _ := startMCPServer(t, oidc.Issuer(), "mcpfather-frontend-test")
	defer cleanupSrv()

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

// TestFrontend_ClientTokenForwarding verifies that sub and email claims are
// forwarded as X-Mcp-Client-Token-Sub / X-Mcp-Client-Token-Email headers.
func TestFrontend_ClientTokenForwarding(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	baseURL, cleanupSrv, mock := startMCPServer(t, oidc.Issuer(), "mcpfather")
	defer cleanupSrv()

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

	if auth := mock.requests[0].Authorization; auth != "" {
		t.Errorf("Token Passthrough Prohibition violated: upstream got '%s'", auth)
	}
}

// TestFrontend_ClientTokenForwardingDisabled verifies that when
// enable_client_token_claim_forward is false, no claim headers are sent.
func TestFrontend_ClientTokenForwardingDisabled(t *testing.T) {
	oidc := startMockOIDCServer(t)
	defer oidc.Close()

	mock := startMockUpstream(okHandler())

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
      enable_client_token_claim_forward: false
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
		mock.Close()
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
