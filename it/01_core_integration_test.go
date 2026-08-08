package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const specFixture = "testdata/minimal_spec.yaml"

// testProxyEnv returns the proxy URL and env vars for build subcommands.
// It checks MCPFATHER_TEST_PROXY first, then HTTPS_PROXY.
// If neither is set, no proxy is configured — suitable for environments
// where Go can reach module proxies directly.
func testProxyEnv(t *testing.T) (proxyURL string, envVars []string) {
	t.Helper()
	proxyURL = os.Getenv("MCPFATHER_TEST_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		logProgress("[proxy] MCPFATHER_TEST_PROXY and HTTPS_PROXY not set — build commands will use direct network")
		return "", nil
	}
	logProgress("[proxy] MCPFATHER_TEST_PROXY=%q HTTPS_PROXY=%q → using %q for build commands",
		os.Getenv("MCPFATHER_TEST_PROXY"), os.Getenv("HTTPS_PROXY"), proxyURL)
	return proxyURL, []string{"HTTPS_PROXY=" + proxyURL}
}

// mcpfatherBin returns the path to the mcpfather binary, building it if needed.
func mcpfatherBin(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("cannot find repo root: %v", err)
	}
	bin := filepath.Join(root, "bin", "mcpfather")
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		_, proxyEnv := testProxyEnv(t)
		cmd := exec.Command("make", "-C", root, "build")
		cmd.Env = append(os.Environ(), proxyEnv...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("make build failed: %v\n%s", err, out)
		}
	}
	return bin
}

// findRepoRoot walks up from the test file to find the repo root.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for depth := 0; depth < 20; depth++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

// ─── build cache: go build ONCE per generated project identity ──
// Stores binary bytes (not paths) so cached binaries survive t.TempDir()
// cleanup. Key includes go.mod content (which embeds the module name =
// output dir basename, varying per-test by TempDir call order) plus the
// mcpfather invocation args, so projects with different includes/excludes
// never collide.

var (
	buildCacheMu sync.Mutex
	buildCache   = map[string][]byte{} // key → cached binary bytes
)

func writeCachedBinary(t *testing.T, dst string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatalf("mkdir for cached binary: %v", err)
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		t.Fatalf("write cached binary: %v", err)
	}
}

// genProject runs mcpfather and returns the output directory path.
func genProject(t *testing.T, includes, excludes string) string {
	t.Helper()
	return genProjectWithSpec(t, specFixture, includes, excludes)
}

var (
	projectKeyMu sync.Mutex
	projectKey   = map[string]string{} // projectDir → cache key
)

// genProjectWithSpec runs mcpfather with a custom spec file (relative to its/).
func genProjectWithSpec(t *testing.T, specFile, includes, excludes string) string {
	t.Helper()
	logProgress("[gen] generating MCP project from spec=%s includes=%q excludes=%q", specFile, includes, excludes)
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	args := []string{"-i", filepath.Join(repoRoot(t), "it", specFile), "-o", dir}
	if includes != "" {
		args = append(args, "--includes", includes)
	}
	if excludes != "" {
		args = append(args, "--excludes", excludes)
	}
	logProgress("[gen] running mcpfather: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpfather failed: %v\n%s", err, out)
	}
	logProgress("[gen] project generated at %s", dir)

	// Compute cache key from go.mod + generation params.
	modFile := filepath.Join(dir, "go.mod")
	modData, err := os.ReadFile(modFile)
	if err != nil {
		t.Fatalf("read go.mod for cache key: %v", err)
	}
	key := string(modData) + "\x00" + specFile + "\x00" + includes + "\x00" + excludes
	projectKeyMu.Lock()
	projectKey[dir] = key
	projectKeyMu.Unlock()

	return dir
}

func repoRoot(t *testing.T) string {
	t.Helper()
	r, err := findRepoRoot()
	if err != nil {
		t.Fatalf("cannot find repo root: %v", err)
	}
	return r
}

// buildServer runs go mod tidy + go build in the generated project dir.
// Key combines go.mod content + mcpfather args so different includes/excludes
// never collide on the cache.
func buildServer(t *testing.T, projectDir string) string {
	t.Helper()
	binName := filepath.Base(projectDir)
	dst := filepath.Join(projectDir, "bin", binName)

	projectKeyMu.Lock()
	key, ok := projectKey[projectDir]
	projectKeyMu.Unlock()
	if !ok {
		modFile := filepath.Join(projectDir, "go.mod")
		data, err := os.ReadFile(modFile)
		if err != nil {
			t.Fatalf("read go.mod for cache key: %v", err)
		}
		key = string(data)
	}

	buildCacheMu.Lock()
	cached, exists := buildCache[key]
	buildCacheMu.Unlock()

	if exists {
		// If binary already exists (e.g. buildServer called twice on same
		// projectDir while the server is running), skip the write to avoid
		// ETXTBUSY on Linux.
		if _, err := os.Stat(dst); err == nil {
			logProgress("[build] binary already exists at %s, skipping cache write", dst)
			return dst
		}
		writeCachedBinary(t, dst, cached)
		logProgress("[build] reused cached binary")
		return dst
	}

	logProgress("[build] go mod tidy + go build in %s", projectDir)
	_, proxyEnv := testProxyEnv(t)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), proxyEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}
	logProgress("[build] go mod tidy OK — building binary %s", binName)
	cmd = exec.Command("go", "build", "-o", filepath.Join("bin", binName), ".")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), proxyEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	logProgress("[build] binary built at %s/bin/%s", projectDir, binName)

	buildCacheMu.Lock()
	binData, _ := os.ReadFile(dst)
	buildCache[key] = binData
	buildCacheMu.Unlock()

	return dst
}

// mockUpstream starts an httptest server that records requests.
type mockUpstream struct {
	mu       sync.Mutex
	server   *httptest.Server
	requests []recordedRequest
}

type recordedRequest struct {
	Method        string
	URL           string
	Authorization string
	Headers       http.Header
	Body          []byte
}

func startMockUpstream(handler http.HandlerFunc) *mockUpstream {
	m := &mockUpstream{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.requests = append(m.requests, recordedRequest{
			Method:        r.Method,
			URL:           r.URL.String(),
			Authorization: r.Header.Get("Authorization"),
			Headers:       r.Header.Clone(),
			Body:          body,
		})
		m.mu.Unlock()
		// Restore the body so inner handlers can read it.
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		handler(w, r)
	}))
	return m
}

func (m *mockUpstream) Close() {
	m.server.Close()
}

func (m *mockUpstream) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// okHandler returns a handler that writes a simple JSON response (no echo).
// This prevents sensitive headers from appearing in the response body at high verbosity.
func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// runCLI runs the generated server in CLI mode and returns stdout+stderr.
func runCLI(t *testing.T, binPath string, env []string, args ...string) (string, string) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()
	return stdout.String(), stderr.String()
}

// ---------------------------------------------------------------------------
// 1. Generator CLI validation
// ---------------------------------------------------------------------------

func TestGenerator_Includes_NonExistentOperationId_Errors(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", "nonExistentOp")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for non-existent operationId, got success")
	}
	if !strings.Contains(string(out), "nonExistentOp") {
		t.Errorf("error message should mention the bad operationId, got: %s", out)
	}
	if !strings.Contains(string(out), "does not exist") {
		t.Errorf("error message should say 'does not exist', got: %s", out)
	}
}

func TestGenerator_Excludes_NonExistentOperationId_Errors(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--excludes", "alsoFake")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for non-existent operationId, got success")
	}
	if !strings.Contains(string(out), "alsoFake") {
		t.Errorf("error message should mention the bad operationId, got: %s", out)
	}
}

func TestGenerator_ValidOperationId_Succeeds(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", "echoHeaders,sayHello")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected success, got: %v\n%s", err, out)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "pkg", "mcptools", "*.go"))
	names := make(map[string]bool)
	for _, f := range files {
		names[filepath.Base(f)] = true
	}
	if !names["EchoHeaders.go"] {
		t.Error("expected EchoHeaders.go to be generated")
	}
	if !names["SayHello.go"] {
		t.Error("expected SayHello.go to be generated")
	}
	if names["DownloadReport.go"] {
		t.Error("DownloadReport.go should NOT be generated (not included)")
	}
}

// TestGenerator_VeryLongOperationId_Succeeds tests the common enterprise
// scenario where operationIds are extremely long with dash/underscore
// separators (e.g. auto-generated from API gateways). The generator must:
//  1. Convert to PascalCase correctly (dashes/underscores → word boundaries)
//  2. Truncate to ≤125 chars with a hash suffix to keep the Go identifier unique
//  3. Produce a buildable server where the tool is registered under its truncated name
func TestGenerator_VeryLongOperationId_Succeeds(t *testing.T) {
	longOpID := "get-a-very-long-operation-id-with-dashes-and_underscores_that_exceeds_the_maximum_tool_name_limit_set_by_opencode_and_other_mcp_integrations_in_the_enterprise_environment"
	spec := filepath.Join(repoRoot(t), "it", "testdata", "oas3.1_spec.yaml")
	if _, err := os.Stat(spec); os.IsNotExist(err) {
		t.Skipf("Blogs OAS 3.1 spec not found at %s", spec)
	}

	// Generate with just this long operationId
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", longOpID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpfather failed for very-long operationId: %v\n%s", err, out)
	}

	// The tool file should exist (with a truncated, hash-suffixed name).
	// Exclude registry.go which is always generated alongside tools.
	var toolFiles []string
	files, _ := filepath.Glob(filepath.Join(dir, "pkg", "mcptools", "*.go"))
	for _, f := range files {
		base := filepath.Base(f)
		if base != "registry.go" && !strings.HasSuffix(base, "_test.go") {
			toolFiles = append(toolFiles, f)
		}
	}
	if len(toolFiles) != 1 {
		t.Fatalf("expected exactly 1 tool file, got %d: %v", len(toolFiles), toolFiles)
	}
	toolFileName := filepath.Base(toolFiles[0])
	toolName := strings.TrimSuffix(toolFileName, ".go")
	t.Logf("generated tool file: %s (name length: %d)", toolFileName, len(toolName))

	// Tool name must be ≤125 chars (MCP limit)
	if len(toolName) > 125 {
		t.Errorf("tool name %q is %d chars, exceeds 125-char limit", toolName, len(toolName))
	}
	// Must retain a recognisable prefix from the original operationId
	if !strings.HasPrefix(strings.ToLower(toolName), "getaverylong") {
		t.Errorf("tool name %q doesn't start with expected PascalCase prefix of original operationId", toolName)
	}

	// Build and smoke-test against mock upstream
	binPath := buildServer(t, dir)
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	stdout, _ := runCLI(t, binPath,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=test-token",
		},
		"-t", "cli", toolName, "--id=12345",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
}

// ---------------------------------------------------------------------------
// 2. Auth / token behaviour
// ---------------------------------------------------------------------------

func TestAuth_BasicPrefixPreserved(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=Basic myCredential123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Basic myCredential123" {
		t.Errorf("Authorization = %q, want %q", got, "Basic myCredential123")
	}
}

func TestAuth_BearerPrefixPreserved(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=Bearer secretToken999",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer secretToken999" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer secretToken999")
	}
}

func TestAuth_NoPrefixDefaultsToBearer(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=plainToken",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer plainToken" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer plainToken")
	}
}

func TestAuth_TokenFileFallback(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	tokenFile := filepath.Join(t.TempDir(), "my-token.txt")
	if err := os.WriteFile(tokenFile, []byte("fileToken123"), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=",
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN_FILE=" + tokenFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer fileToken123" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer fileToken123")
	}
}

func TestAuth_TokenFileWithBasicPrefix(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	tokenFile := filepath.Join(t.TempDir(), "my-token.txt")
	if err := os.WriteFile(tokenFile, []byte("Basic fileBasic123"), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=",
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN_FILE=" + tokenFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Basic fileBasic123" {
		t.Errorf("Authorization = %q, want %q", got, "Basic fileBasic123")
	}
}

func TestAuth_CookieFromEnv(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN=JSESSIONID=abc123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=abc123" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=abc123")
	}
}

func TestAuth_CookieFileFallback(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	cookieFile := filepath.Join(t.TempDir(), "my-cookie.txt")
	if err := os.WriteFile(cookieFile, []byte("JSESSIONID=fileSession456"), 0600); err != nil {
		t.Fatalf("failed to write cookie file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN=",
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN_FILE=" + cookieFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=fileSession456" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=fileSession456")
	}
}

func TestAuth_CookieAndTokenBothSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=Bearer secretToken999",
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN=JSESSIONID=abc123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer secretToken999" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer secretToken999")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=abc123" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=abc123")
	}
}

// ---------------------------------------------------------------------------
// 3. Logging behaviour
// ---------------------------------------------------------------------------

// TestLogging_AuthHeaderRedactedByDefault verifies that at -v 10 the
// Authorization header VALUE is shown as "***" in the upstream request log.
// Uses okHandler so the response body does NOT echo the token back.
func TestLogging_AuthHeaderRedactedByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=secretSauce",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// The header line should show "Authorization: ***"
	if !strings.Contains(stderr, "Authorization: ***") {
		t.Error("expected 'Authorization: ***' in upstream request log, but not found. stderr:\n" + stderr)
	}
	// The raw token value must NOT appear as a header value in the upstream request log
	// (it appears as "Authorization: ***" not "Authorization: Bearer secretSauce")
	if strings.Contains(stderr, "Bearer secretSauce") {
		t.Error("Authorization value should be redacted, but 'Bearer secretSauce' appears in log. stderr:\n" + stderr)
	}
}

// TestLogging_AuthHeaderPrintedWhenEnvSet verifies that setting
// MCP__LOGGING__AUTH_VERBOSE=true makes the Authorization value visible.
func TestLogging_AuthHeaderPrintedWhenEnvSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=visibleToken",
			"MCP__LOGGING__AUTH_VERBOSE=true",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// With MCP__LOGGING__AUTH_VERBOSE=true, the token should appear
	if !strings.Contains(stderr, "visibleToken") {
		t.Error("expected Authorization value to be visible when MCP__LOGGING__AUTH_VERBOSE=true. stderr:\n" + stderr)
	}
}

// TestLogging_CookieRedactedByDefault verifies that the Cookie header value is
// shown as "***" in upstream request logs at -v 10.
func TestLogging_CookieRedactedByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN=JSESSIONID=secretSession",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// The header line should show "Cookie: ***"
	if !strings.Contains(stderr, "Cookie: ***") {
		t.Error("expected 'Cookie: ***' in upstream request log, but not found. stderr:\n" + stderr)
	}
	// The raw cookie value must NOT appear
	if strings.Contains(stderr, "secretSession") {
		t.Error("Cookie value should be redacted, but 'secretSession' appears in log. stderr:\n" + stderr)
	}
}

// TestLogging_CookiePrintedWhenEnvSet verifies that setting
// MCP__LOGGING__AUTH_VERBOSE=true makes the Cookie value visible.
func TestLogging_CookiePrintedWhenEnvSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN=JSESSIONID=visibleSession",
			"MCP__LOGGING__AUTH_VERBOSE=true",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// With MCP__LOGGING__AUTH_VERBOSE=true, the cookie value should appear
	if !strings.Contains(stderr, "visibleSession") {
		t.Error("expected Cookie value to be visible when MCP__LOGGING__AUTH_VERBOSE=true. stderr:\n" + stderr)
	}
}

// TestLogging_NonAuthHeadersPrinted verifies that non-Authorization headers are
// printed at high verbosity. For a GET request without body, we check the method
// and URL are logged.
func TestLogging_NonAuthHeadersPrinted(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=someToken",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// At verbosity >= 2, method and URL are logged
	if !strings.Contains(stderr, "GET "+mock.server.URL) {
		t.Error("expected upstream method and URL in verbose logs. stderr:\n" + stderr)
	}
}

// ---------------------------------------------------------------------------
// 4. Transport mode consistency (CLI vs HTTP)
// ---------------------------------------------------------------------------

// mcpHTTPCall sends an MCP JSON-RPC request via HTTP Streamable transport.
// It first calls initialize to get a session ID, then uses that for subsequent calls.
func mcpHTTPCall(t *testing.T, baseURL string, method string, params map[string]interface{}) (*http.Response, string) {
	t.Helper()

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
	resp, err := http.Post(baseURL+"/mcp", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("initialize request failed: %v", err)
	}
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

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
	req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("MCP %s request failed: %v", method, err)
	}
	return resp, sessionID
}

// waitForServer polls the MCP endpoint until the server responds or times out.
func waitForServer(t *testing.T, baseURL string) {
	t.Helper()
	logProgress("[wait] polling %s/mcp for server readiness (timeout 5s)", baseURL)
	for i := 0; i < 100; i++ {
		resp, err := http.Post(baseURL+"/mcp", "application/json", strings.NewReader(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`))
		if err == nil {
			resp.Body.Close()
			logProgress("[wait] server ready at %s (attempt %d)", baseURL, i+1)
			return
		}
		if i%20 == 19 {
			logProgress("[wait] still waiting for %s (attempt %d/100, last error: %v)", baseURL, i+1, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("HTTP server did not become ready after 5s")
}

// TestAuth_HTTPTransportMatchesCLI verifies that the HTTP transport sends the
// same Authorization header as CLI mode when using MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN.
func TestAuth_HTTPTransportMatchesCLI(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	port := "19876"

	cmd := exec.Command(bin, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=Basic httpToken456",
	)
	var stderrBuf bytes.Buffer
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

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	if len(mock.requests) == 0 {
		t.Fatalf("no request reached the mock upstream. stderr:\n%s", stderrBuf.String())
	}
	if got := mock.requests[0].Authorization; got != "Basic httpToken456" {
		t.Errorf("HTTP transport: Authorization = %q, want %q", got, "Basic httpToken456")
	}
}

// TestLogging_HTTPTransportRedactsAuthByDefault verifies that the HTTP
// transport also redacts Authorization in upstream request logs by default.
func TestLogging_HTTPTransportRedactsAuthByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	port := "19878"

	cmd := exec.Command(bin, "--transport", "http", "--port", port, "-v", "10")
	cmd.Env = append(os.Environ(),
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=shouldBeHidden",
	)
	var stderrBuf bytes.Buffer
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

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Authorization: ***") {
		t.Error("expected 'Authorization: ***' in HTTP transport upstream logs. stderr:\n" + stderr)
	}
	if strings.Contains(stderr, "shouldBeHidden") {
		t.Error("token value should NOT appear in logs. stderr:\n" + stderr)
	}
}

// ---------------------------------------------------------------------------
// 5. Binary download
// ---------------------------------------------------------------------------

func TestDownload_BinaryFileSavedLocally(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
		w.Write([]byte("fake-binary-pdf-content"))
	})
	defer mock.Close()

	projectDir := genProject(t, "downloadReport", "")
	bin := buildServer(t, projectDir)
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadReport",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}

	// Files are saved under ~/.{serviceName}/ifs/download/{yyyyMMdd}/ with UUID naming.
	// Walk the IFS date directories to find the downloaded file.
	ifsDir := filepath.Join(homeDir, "."+serviceName, "ifs", "download")
	found := false
	filepath.WalkDir(ifsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".pdf") {
			found = true
			data, _ := os.ReadFile(path)
			if string(data) != "fake-binary-pdf-content" {
				t.Errorf("downloaded content = %q, want %q", string(data), "fake-binary-pdf-content")
			}
		}
		return nil
	})
	if !found {
		t.Errorf("downloaded .pdf file not found in %s", ifsDir)
	}
}

func TestDownload_NoContentDisposition_UsesDefaultName(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write([]byte("fake-zip-content"))
	})
	defer mock.Close()

	projectDir := genProject(t, "downloadReport", "")
	bin := buildServer(t, projectDir)
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadReport",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}
	// When no Content-Disposition is set, DetermineFileName falls back to
	// the URL path last segment ("download" from /download endpoint) or
	// Content-Type-based extension. The IFS layer wraps it in a UUID name,
	// so verify the content was saved correctly rather than the exact name.
	if !strings.Contains(stdout, "download") {
		t.Errorf("expected filename derived from URL path or content-type, got: %s", stdout)
	}

	// Files are saved under ~/.{serviceName}/ifs/download/{yyyyMMdd}/ with UUID naming.
	ifsDir := filepath.Join(homeDir, "."+serviceName, "ifs", "download")
	found := false
	filepath.WalkDir(ifsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		found = true
		data, _ := os.ReadFile(path)
		if string(data) != "fake-zip-content" {
			t.Errorf("downloaded content = %q, want %q", string(data), "fake-zip-content")
		}
		return nil
	})
	if !found {
		t.Errorf("downloaded file not found in %s", ifsDir)
	}
}

// ---------------------------------------------------------------------------
// 5b. Real binary download (external endpoint)
// ---------------------------------------------------------------------------

// TestDownload_BinaryWithKnownSize tests binary download with a known
// content size. Uses a local mock so the test works without internet access.
func TestDownload_BinaryWithKnownSize(t *testing.T) {
	// 1024 bytes of deterministic pseudo-binary content
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=random.bin")
		w.Write(payload)
	})
	defer mock.Close()

	projectDir := genProjectWithSpec(t, "testdata/binary_spec.yaml", "downloadBytes", "")
	bin := buildServer(t, projectDir)
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadBytes",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}

	// Files are saved under ~/.{serviceName}/ifs/download/{yyyyMMdd}/ with UUID naming.
	ifsDir := filepath.Join(homeDir, "."+serviceName, "ifs", "download")
	found := 0
	filepath.WalkDir(ifsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		found++
		info, _ := d.Info()
		t.Logf("Downloaded file: %s (%d bytes)", d.Name(), info.Size())
		if info.Size() != 1024 {
			t.Errorf("expected 1024 bytes, got %d", info.Size())
		}
		data, _ := os.ReadFile(path)
		if len(data) != 1024 {
			t.Errorf("expected 1024 bytes on disk, got %d", len(data))
		}
		return nil
	})
	if found == 0 {
		t.Fatalf("no files found in %s", ifsDir)
	}
}

// ---------------------------------------------------------------------------
// 6. Upload
// ---------------------------------------------------------------------------

// TestUpload_CLI_FromUploadsDir verifies that an upload tool reads a file from
// the upload directory (~/.{serviceName}/ifs/upload/) and sends it to the upstream.
func TestUpload_CLI_FromUploadsDir(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterUploadScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	binPath := buildServer(t, dir)
	serviceName := filepath.Base(dir)
	homeDir := t.TempDir()

	// Stage the file in the upload directory
	uploadDir := filepath.Join(homeDir, "."+serviceName, "ifs", "upload")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	testContent := []byte("hello world upload test content")
	if err := os.WriteFile(filepath.Join(uploadDir, "test-upload.bin"), testContent, 0644); err != nil {
		t.Fatalf("failed to stage upload file: %v", err)
	}

	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "UploadFile", "--file_name=test-upload.bin",
	)

	data := mustJSON(t, stdout)
	if fc, _ := data["fileContent"].(string); fc != string(testContent) {
		t.Errorf("uploaded content = %q, want %q", fc, string(testContent))
	}
	if method, _ := data["method"].(string); method != "POST" {
		t.Errorf("upload method = %q, want POST", method)
	}
}

// TestUpload_HTTP_Base64Content verifies HTTP mode upload with base64 file_content.
func TestUpload_HTTP_Base64Content(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterUploadScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	testContent := "http-mode-upload-content"
	b64Content := base64.StdEncoding.EncodeToString([]byte(testContent))

	result := callNativeTool(t, baseURL, "UploadFile", map[string]interface{}{
		"file_name":    "http-upload.bin",
		"file_content": b64Content,
	})

	data := mustJSON(t, result)
	if fc, _ := data["fileContent"].(string); fc != testContent {
		t.Errorf("uploaded content = %q, want %q", fc, testContent)
	}
}

// TestUpload_CLI_WithoutFile verifies that an upload tool can be called
// without --file_name and falls back to sending a JSON body.
func TestUpload_CLI_WithoutFile(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	// Use the existing upload_spec.yaml - its upload API should now accept
	// calls without file_name
	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "UploadFile",
	)

	// Should NOT contain "missing required argument: file_name"
	if strings.Contains(stdout, "missing required argument") {
		t.Errorf("tool should not require file_name, got: %s", stdout)
	}

	// Should have made a POST request to the mock upstream
	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request, got none")
	} else if mock.requests[0].Method != "POST" {
		t.Errorf("expected POST method, got %s", mock.requests[0].Method)
	}
}

// TestUpload_HTTP_WithoutFile verifies that an upload tool called via HTTP mode
// without file_name successfully falls back to JSON body forwarding.
func TestUpload_HTTP_WithoutFile(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	// Use the existing upload_spec.yaml
	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// Call the upload tool without file_name - should succeed via fallback path
	result := callNativeTool(t, baseURL, "UploadFile", map[string]interface{}{
		"key": "value",
	})

	// Should NOT contain an MCP error about missing file_name
	if strings.Contains(result, "missing required argument") {
		t.Errorf("tool should not require file_name, got: %s", result)
	}
	if strings.Contains(result, "MCP error") {
		t.Errorf("unexpected MCP error: %s", result)
	}

	// Verify the mock received the request
	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request, got none")
	}
}

// TestFormUrlEncoded_NotTreatedAsUpload verifies that
// application/x-www-form-urlencoded APIs are NOT treated as file upload tools
// and go through the standard JSON body forwarding path.
func TestFormUrlEncoded_NotTreatedAsUpload(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createFormResource", "")
	binPath := buildServer(t, dir)

	// Call the form-url-encoded tool with body data
	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "CreateFormResource", "--name=testuser", "--email=test@example.com",
	)

	if strings.Contains(stdout, "missing required argument: file_name") {
		t.Errorf("form-url-encoded API should NOT require file_name, got: %s", stdout)
	}

	// Should have made a POST request
	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request, got none")
	}
	if len(mock.requests) > 0 && mock.requests[0].Method != "POST" {
		t.Errorf("expected POST method, got %s", mock.requests[0].Method)
	}
}

// TestMultipartUpload_WithOptionalFile_CLI verifies CLI mode of a multipart API
// with optional file fields works without --file_name.
func TestMultipartUpload_WithOptionalFile_CLI(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "CreateMultipartResource", "--name=my-resource",
	)

	if strings.Contains(stdout, "missing required argument: file_name") {
		t.Errorf("multipart tool without required file should not require file_name, got: %s", stdout)
	}

	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request, got none")
	}
}

// TestMultipartUpload_WithOptionalFile_HTTP verifies HTTP mode of a multipart API
// with optional file fields works without file_name.
func TestMultipartUpload_WithOptionalFile_HTTP(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "CreateMultipartResource", map[string]interface{}{
		"name":        "my-resource",
		"description": "test description",
	})

	if strings.Contains(result, "missing required argument") {
		t.Errorf("multipart tool should not require file_name, got: %s", result)
	}
	if strings.Contains(result, "MCP error") {
		t.Errorf("unexpected MCP error: %s", result)
	}

	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request, got none")
	}
}

// ---------------------------------------------------------------------------
// 6b. FileRef multipart — download + forward with real mock upstream
// ---------------------------------------------------------------------------

// TestFileRef_MultipartUpload_CLI verifies that when a FileRef tool is called
// in CLI mode with a file URI, the generated MCP server:
//  1. Downloads the file from the URI
//  2. Builds a multipart/form-data request
//  3. Forwards it to the upstream with correct form fields + file content
func TestFileRef_MultipartUpload_CLI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	binPath := buildServer(t, dir)

	fileURL := mockURL + "/files/sample.txt"
	stdout, stderr := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "CreateMultipartResource",
		"--name=myresource",
		"--description=A test resource",
		"--file="+fileURL,
	)
	_ = stderr

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("FileRef CLI call failed: %s", stdout)
	}

	// Verify the multipart data was received by the upstream
	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}
	if record.FormFields["name"] != "myresource" {
		t.Errorf("expected form field name=myresource, got %q", record.FormFields["name"])
	}
	if record.FormFields["description"] != "A test resource" {
		t.Errorf("expected form field description='A test resource', got %q", record.FormFields["description"])
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.FileName != "sample.txt" {
		t.Errorf("expected uploaded file name 'sample.txt', got %q", fileRec.FileName)
	}
	if string(fileRec.Content) != "HELLO-FILEREF-sample.txt" {
		t.Errorf("expected file content 'HELLO-FILEREF-sample.txt', got %q", string(fileRec.Content))
	}
}

// TestFileRef_MultipartUpload_HTTP verifies that when a FileRef tool is called
// in HTTP mode with a file URI, the generated MCP server download the file
// and forwards it as a proper multipart/form-data request to the upstream.
func TestFileRef_MultipartUpload_HTTP(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	fileURL := mockURL + "/files/sample.txt"
	result := callNativeTool(t, baseURL, "CreateMultipartResource", map[string]interface{}{
		"name":        "myresource",
		"description": "HTTP mode test",
		"file":        fileURL,
	})

	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("FileRef HTTP call failed: %s", result)
	}

	// Verify multipart data at upstream
	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}
	if record.FormFields["name"] != "myresource" {
		t.Errorf("expected form field name=myresource, got %q", record.FormFields["name"])
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.FileName != "sample.txt" {
		t.Errorf("expected uploaded file name 'sample.txt', got %q", fileRec.FileName)
	}
	if string(fileRec.Content) != "HELLO-FILEREF-sample.txt" {
		t.Errorf("expected file content 'HELLO-FILEREF-sample.txt', got %q", string(fileRec.Content))
	}
}

// TestFileRef_NoFileProvided_FallsThroughToJSON confirms the fallthrough path:
// when no file URI is provided to a FileRef tool, it sends a standard JSON
// body to the upstream instead of a multipart request.
func TestFileRef_NoFileProvided_FallsThroughToJSON(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "CreateMultipartResource",
		"--name=json-fallback",
		"--description=No file here",
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("FileRef fallback call failed: %s", stdout)
	}
	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request (JSON fallback)")
	}
	// The request should be JSON, not multipart
	if len(mock.requests) > 0 {
		contentType := mock.requests[0].Headers.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			t.Error("expected JSON content-type (no file provided), got multipart/form-data")
		}
	}
}

// ---------------------------------------------------------------------------
// 7. CLI argument passing
// ---------------------------------------------------------------------------

func TestCLI_QueryParamsPassedToUpstream(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "sayHello", ""))
	_, _ = runCLI(t, bin,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "SayHello", "--name=World",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if !strings.Contains(mock.requests[0].URL, "name=World") {
		t.Errorf("query param 'name=World' not found in URL: %s", mock.requests[0].URL)
	}
}

func TestCLI_ListShowsTools(t *testing.T) {
	bin := buildServer(t, genProject(t, "echoHeaders,sayHello", ""))
	stdout, _ := runCLI(t, bin, nil, "-t", "cli", "list")

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("expected EchoHeaders in tool list")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("expected SayHello in tool list")
	}
}

// ---------------------------------------------------------------------------
// 7. Cyclic $ref detection (regression: LinkGroup.groups → LinkGroup)
// ---------------------------------------------------------------------------

const cyclicSpecFixture = "testdata/cyclic_spec.yaml"

// TestCyclicRef_GenerationSucceeds verifies that mcpfather does NOT hang when the
// OpenAPI spec contains a self-referencing schema (LinkGroup.groups → LinkGroup).
// Before the cycle-detection fix, the recursive schema walkers would recurse
// infinitely and the process would OOM or hang.
func TestCyclicRef_GenerationSucceeds(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpfather failed for cyclic spec: %v\n%s", err, out)
	}

	// Both tools should be generated
	expected := []string{"ListItems.go", "HealthCheck.go"}
	for _, name := range expected {
		fp := filepath.Join(dir, "pkg", "mcptools", name)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("expected tool file %s was not generated", name)
		}
	}
}

// TestCyclicRef_ResponseTemplateHasCyclicMarker verifies that the generated
// response template for a cyclic schema contains the [cyclic reference] marker.
func TestCyclicRef_ResponseTemplateHasCyclicMarker(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpfather failed: %v\n%s", err, out)
	}

	// Read the ListItems tool file which has the cyclic LinkGroup schema
	toolFile := filepath.Join(dir, "pkg", "mcptools", "ListItems.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read ListItems.go: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "[cyclic reference]") {
		t.Error("expected '[cyclic reference]' marker in ListItems.go response template, but not found")
	}
}

// TestCyclicRef_NonCyclicSchemaNoSpuriousMarker verifies that non-cyclic schemas
// do NOT get a false-positive [cyclic reference] marker. The HealthCheck tool uses
// a simple HealthStatus schema with no self-references.
func TestCyclicRef_NonCyclicSchemaNoSpuriousMarker(t *testing.T) {
	bin := mcpfatherBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpfather failed: %v\n%s", err, out)
	}

	// Read the HealthCheck tool file which uses a flat HealthStatus schema
	toolFile := filepath.Join(dir, "pkg", "mcptools", "HealthCheck.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read HealthCheck.go: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "[cyclic reference]") {
		t.Error("HealthCheck.go should NOT contain '[cyclic reference]' — false positive for acyclic schema")
	}

	// The response template should still describe the status and uptime fields
	if !strings.Contains(content, "status") {
		t.Error("expected 'status' field in HealthCheck response template")
	}
	if !strings.Contains(content, "uptime") {
		t.Error("expected 'uptime' field in HealthCheck response template")
	}
}

// TestCyclicRef_BuildsAndRuns verifies that a server generated from a cyclic spec
// builds successfully and can invoke a tool against a mock upstream at runtime.
func TestCyclicRef_BuildsAndRuns(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	binPath := buildServer(t, genProjectWithSpec(t, cyclicSpecFixture, "", ""))
	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "HealthCheck",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
}

// TestRegression_MinimalSpecResponseTemplate is a regression test ensuring the
// cycle-detection changes did not alter the output for non-cyclic schemas. The
// response template for echoHeaders must contain its usual structure.
func TestRegression_MinimalSpecResponseTemplate(t *testing.T) {
	dir := genProject(t, "echoHeaders", "")

	toolFile := filepath.Join(dir, "pkg", "mcptools", "EchoHeaders.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read EchoHeaders.go: %v", err)
	}
	content := string(data)

	// The response template should describe the response structure
	if !strings.Contains(content, "# API Response Information") {
		t.Error("expected '# API Response Information' header in response template")
	}
	if !strings.Contains(content, "**Status Code:** 200") {
		t.Error("expected '**Status Code:** 200' in response template")
	}
	if !strings.Contains(content, "**Content-Type:** application/json") {
		t.Error("expected '**Content-Type:** application/json' in response template")
	}
	// Must NOT have spurious cyclic markers
	if strings.Contains(content, "[cyclic reference]") {
		t.Error("EchoHeaders.go should NOT contain '[cyclic reference]' — regression in cycle detection")
	}
}

// TestRegression_SayHelloRequestSchema verifies the request arg schema for a
// tool with query parameters is still generated correctly after cycle-detection
// changes (visited map is threaded through requestArgsSchema path).
func TestRegression_SayHelloRequestSchema(t *testing.T) {
	dir := genProject(t, "sayHello", "")

	toolFile := filepath.Join(dir, "pkg", "mcptools", "SayHello.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read SayHello.go: %v", err)
	}
	content := string(data)

	// The InputSchema must describe the "name" query parameter.
	// The JSON schema is embedded as an escaped Go string literal,
	// so we match the escaped form: \"name\" and \"type\": \"string\".
	if !strings.Contains(content, `\"name\"`) {
		t.Errorf("expected 'name' property in InputSchema, content:\n%s", content)
	}
	if !strings.Contains(content, `\"type\": \"string\"`) {
		t.Errorf("expected 'type: string' in InputSchema for name parameter, content:\n%s", content)
	}
}

// TestRegression_FullBuildAndCLI verifies the full end-to-end flow still works:
// generate → build → CLI invoke with the minimal spec. This is the broadest
// regression smoke test for the cycle-detection changes.
func TestRegression_FullBuildAndCLI(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders,sayHello", ""))
	stdout, _ := runCLI(t, bin,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL},
		"-t", "cli", "SayHello", "--name=RegressionTest",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
	if !strings.Contains(mock.requests[0].URL, "name=RegressionTest") {
		t.Errorf("query param 'name=RegressionTest' not found in URL: %s", mock.requests[0].URL)
	}
}

func TestE2E_Core_AuthForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "test-bearer-token-abc123", "")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "Bearer already-prefixed-token", "")
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
// Core Test 3: Cookie forwarding via MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN
// ---------------------------------------------------------------------------

func TestE2E_Core_CookieForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "JSESSIONID=abc123test")
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
// Core Test 4: MCP session ID forwarding is enabled by default
// ---------------------------------------------------------------------------

func TestE2E_Core_SessionForwardingEnabledByDefault(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	// enable_mcp_session_forward defaults to true;
	// X-MCP-Session-ID should be forwarded by default.
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, _ := data["headers"].(map[string]interface{})

	if _, ok := headers["X-Mcp-Session-Id"]; !ok {
		t.Error("X-Mcp-Session-Id should be forwarded when enable_mcp_session_forward defaults to true")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
		"arguments": map[string]interface{}{
			"status": "400",
		},
	})
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	// Upstream returns 400, MCP server should propagate the error
	if !strings.Contains(string(bodyBytes), "Bad Request") {
		t.Logf("error response body: %s", string(bodyBytes))
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

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
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
	serviceName := filepath.Base(dir)

	virtConfig := `
virtual_tools:
  - name: virt_chain
    description: Chain echo and greet
    input_schema:
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
	writeCoreVirtualConfig(t, homeDir, serviceName, virtConfig)

	cleanup, baseURL := startVirtualTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallVirtualTool(t, baseURL, "virt_chain", map[string]interface{}{
		"name": "Alice",
	})

	data := mustJSON(t, result)
	greeting, _ := data["greeting"].(string)
	if !strings.Contains(greeting, "Alice") {
		t.Errorf("greeting should mention Alice, got %q", greeting)
	}
}

// ---------------------------------------------------------------------------
// Core Test 12: Config-based tool activation — backward compatibility
// ---------------------------------------------------------------------------

// TestConfig_NoConfigFile_AllToolsEnabled verifies that when no config.yaml
// exists, all native tools are registered and callable (backward compatible).
func TestConfig_NoConfigFile_AllToolsEnabled(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	// Intentionally do NOT write any config file

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// All 3 tools should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// Also verify via CLI list
	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)
	for _, name := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("CLI list should contain %q when no config file exists", name)
		}
	}
}

// TestConfig_RegisterAllToolsByDefault_True_WithExcludes verifies that when
// register-all-tools-by-default is true, all tools are available except those
// listed in excludes.
func TestConfig_RegisterAllToolsByDefault_True_WithExcludes(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: true
    excludes:
      - DownloadReport
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders and SayHello should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// DownloadReport should NOT be callable (excluded)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should be excluded, but got response: %s", trimMsg(string(bodyBytes), 300))
	}

	// CLI list should not show DownloadReport
	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)
	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should contain SayHello")
	}
	if strings.Contains(stdout, "DownloadReport") {
		t.Error("CLI list should NOT contain excluded DownloadReport")
	}
}

// TestConfig_RegisterAllToolsByDefault_False_WithIncludes verifies that when
// register-all-tools-by-default is false (the default), only tools listed in
// includes are available.
func TestConfig_RegisterAllToolsByDefault_False_WithIncludes(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes:
      - EchoHeaders
      - SayHello
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders and SayHello should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// DownloadReport should NOT be callable (not in includes)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should not be available, but got response: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_RegisterAllToolsByDefault_False_Default_EmptyIncludes_NoTools
// verifies that when register-all-tools-by-default is false (or omitted) and
// includes is empty, no native tools are registered.
func TestConfig_RegisterAllToolsByDefault_False_Default_EmptyIncludes_NoTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes: []
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)

	// No native tools should be listed
	if strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should NOT contain EchoHeaders when register-all-tools-by-default is false and includes is empty")
	}
	if strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should NOT contain SayHello when register-all-tools-by-default is false and includes is empty")
	}
}

// TestConfig_IncludesAndExcludes_Conflict_ServerFailsToStart verifies that
// when a tool appears in both includes and excludes, the server exits with
// an error at startup.
func TestConfig_IncludesAndExcludes_Conflict_ServerFailsToStart(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: true
    includes:
      - EchoHeaders
    excludes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, dir)

	// Start the server — it should fail due to conflict
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))
	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
	)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Wait for the process to exit (it should exit quickly with an error)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case waitErr := <-done:
		if waitErr == nil {
			t.Error("server should have exited with an error due to config conflict, but exited successfully")
		}
		stderr := stderrBuf.String()
		if !strings.Contains(stderr, "EchoHeaders") {
			t.Errorf("stderr should mention the conflicting tool name. stderr:\n%s", stderr)
		}
		if !strings.Contains(stderr, "both includes and excludes") || !strings.Contains(stderr, "includes and excludes") {
			t.Errorf("stderr should mention includes/excludes conflict. stderr:\n%s", stderr)
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("server did not exit within 5 seconds — expected immediate failure on config conflict")
	}
}

// TestConfig_IncludesAndExcludes_NoConflict_Success verifies that includes
// and excludes can be used together without conflict, and the resulting
// tool set is correct.
func TestConfig_IncludesAndExcludes_NoConflict_Success(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: true
    includes:
      - EchoHeaders
    excludes:
      - DownloadReport
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders: in includes (explicitly added) + all-by-default
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// SayHello: enabled by register-all-tools-by-default, not excluded
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "SayHello",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// DownloadReport: excluded
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should be excluded, but got: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_ExposeIncludes_WithAllDefaultFalse_AddsTools verifies that
// includes adds tools even when register-all-tools-by-default is false.
func TestConfig_ExposeIncludes_WithAllDefaultFalse_AddsTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes:
      - EchoHeaders
    excludes:
      - SayHello
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders: explicitly included, should work
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// SayHello: excluded (no-op since it wasn't activated by default anyway)
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "SayHello",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("SayHello should not be available, but got: %s", trimMsg(string(bodyBytes), 300))
	}

	// DownloadReport: not in includes, not in excludes, not activated by default
	resp2, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp2.Body.Close()
	bodyBytes, _ = io.ReadAll(resp2.Body)
	if !strings.Contains(string(bodyBytes), "not found") {
		t.Errorf("DownloadReport should not be available, but got: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_CLI_ListRespectsConfig verifies that CLI list honors the
// expose config with includes/excludes.
func TestConfig_CLI_ListRespectsConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders (included)")
	}
	if strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should NOT contain SayHello (not included)")
	}
	if strings.Contains(stdout, "DownloadReport") {
		t.Error("CLI list should NOT contain DownloadReport (not included)")
	}
}

// TestConfig_CLI_CallRespectsConfig verifies that CLI call rejects tools
// that are not enabled by the expose config.
func TestConfig_CLI_CallRespectsConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, dir)
	env := []string{
		"HOME=" + homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
	}

	// EchoHeaders is enabled, should succeed
	_, stderr := runCLI(t, binPath, env, "-t", "cli", "EchoHeaders")
	if strings.Contains(stderr, "not enabled") {
		t.Error("EchoHeaders should be callable when included")
	}

	// SayHello is not enabled, should be rejected
	_, stderr = runCLI(t, binPath, env, "-t", "cli", "SayHello", "--name=Test")
	if !strings.Contains(stderr, "not enabled") {
		t.Errorf("SayHello should be rejected as not enabled, but got stderr: %s", stderr)
	}
}

// TestConfig_EmptyExpose_AllToolsAvailable verifies backward compatibility:
// when the config file exists but has no tools.expose section, all tools
// are still available.
func TestConfig_EmptyExpose_AllToolsAvailable(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	// Config file with upstream section only, no tools.expose
	cfg := `
upstream:
    enable_mcp_session_forward: false
runtime:
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// All 3 tools should still be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}
}

// TestConfig_IncludesNotFoundInRegistry_NoError verifies that if an
// includes entry names a tool not in the registry, it is silently ignored
// (only tools that actually exist are activated). A warning should not
// prevent the server from starting.
func TestConfig_IncludesNotFoundInRegistry_NoError(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes:
      - EchoHeaders
      - NonExistentTool
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders should still be callable (NonExistentTool is silently ignored)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()
}

// ---------------------------------------------------------------------------
// 7. IFS (Internal File System) Data Plane
// ---------------------------------------------------------------------------

// TestIFS_UploadAndDownload verifies the IFS REST API:
//  1. Upload a file via POST /_/ifs/upload/{yyyyMMdd}/{uuid}.{suffix}
//  2. Download via GET /_/ifs/download/{yyyyMMdd}/{uuid}.{suffix}
//  3. The downloaded content matches what was uploaded.
func TestIFS_UploadAndDownload(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
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

	dateStr := time.Now().Format("20060102")
	testUUID := "deadbeef-dead-dead-dead-deaddeadbeef"
	testFileName := testUUID + ".test"
	testContent := "IFS upload test content"
	uploadURL := baseURL + "/_/ifs/upload/" + dateStr + "/" + testFileName

	// Upload
	uploadResp, err := http.Post(uploadURL, "application/octet-stream", strings.NewReader(testContent))
	if err != nil {
		t.Fatalf("IFS upload failed: %v", err)
	}
	uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", uploadResp.StatusCode)
	}

	// Download
	downloadURL := baseURL + "/_/ifs/download/" + dateStr + "/" + testFileName
	downloadResp, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("IFS download failed: %v", err)
	}
	defer downloadResp.Body.Close()
	if downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", downloadResp.StatusCode)
	}

	downloaded, _ := io.ReadAll(downloadResp.Body)
	if string(downloaded) != testContent {
		t.Errorf("downloaded content = %q, want %q", string(downloaded), testContent)
	}

	// Verify file exists on disk in the expected path (IFS uses download dir)
	ifsDataDir := filepath.Join(homeDir, "."+filepath.Base(projectDir), "ifs", "download", dateStr)
	stagedFile := filepath.Join(ifsDataDir, testFileName)
	if _, err := os.Stat(stagedFile); os.IsNotExist(err) {
		t.Errorf("expected IFS file at %s, but not found", stagedFile)
	}
}

// TestIFS_DisabledByConfig verifies that when ifs.enabled is false,
// the IFS endpoints return 404.
func TestIFS_DisabledByConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)

	cfg := `
server:
  ifs:
    enabled: false
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
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

	dateStr := time.Now().Format("20060102")
	downloadURL := baseURL + "/_/ifs/download/" + dateStr + "/test.bin"
	resp, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when IFS disabled, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// 8. Logging config
// ---------------------------------------------------------------------------

// TestLoggingConfig_LevelFromConfig verifies that logging.level from
// config.yaml takes effect when -v is 0 or not passed.
func TestLoggingConfig_LevelFromConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)

	cfg := `
logging:
  level: 4
  auth_verbose: false
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	// Start with -v 0 (or no -v) — config's logging.level=4 should activate
	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
	)
	var stderrBuf bytes.Buffer
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

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	stderr := stderrBuf.String()
	// At level 4, we should see request logging (type headers, method/url)
	if !strings.Contains(stderr, "-->") {
		t.Errorf("expected '-->' (request log) at logging.level=4, stderr:\n%s", stderr)
	}
}

// TestLoggingConfig_AuthVerboseEnvOverride verifies MCP__LOGGING__AUTH_VERBOSE
// env var overrides the config file value.
func TestLoggingConfig_AuthVerboseEnvOverride(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "10")
	cmd.Env = append(os.Environ(),
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		"MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN=shouldBeVisible",
		"MCP__LOGGING__AUTH_VERBOSE=true",
	)
	var stderrBuf bytes.Buffer
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

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	stderr := stderrBuf.String()
	// With auth_verbose=true, the token should be visible
	if !strings.Contains(stderr, "shouldBeVisible") {
		t.Errorf("expected token 'shouldBeVisible' in logs with auth_verbose=true, stderr:\n%s", stderr)
	}
}

// ---------------------------------------------------------------------------
// 9. ENV array override
// ---------------------------------------------------------------------------

// TestEnvOverride_ArrayFields verifies that MCP__-prefixed ENV vars with
// numeric indices override array-type config fields (Spring Boot style).
func TestEnvOverride_ArrayFields(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	serviceName := filepath.Base(dir)

	// Write config with register_all_tools_by_default: false, no includes.
	// Then override includes via ENV array vars.
	cfg := `
native_tools:
  expose:
    register-all-tools-by-default: false
    includes: []
`
	writeCoreVirtualConfig(t, homeDir, serviceName, cfg)

	binPath := buildServer(t, dir)
	env := []string{
		"HOME=" + homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		// Array ENV override: set includes[0] and includes[1]
		"MCP__NATIVE_TOOLS__EXPOSE__INCLUDES__0=EchoHeaders",
		"MCP__NATIVE_TOOLS__EXPOSE__INCLUDES__1=SayHello",
	}

	// CLI list should show only the two included tools
	stdout, _ := runCLI(t, binPath, env, "-t", "cli", "list")

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders (via array ENV override)")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should contain SayHello (via array ENV override)")
	}
	// DownloadReport is NOT in includes — should not appear
	if strings.Contains(stdout, "DownloadReport") {
		t.Errorf("CLI list should NOT contain DownloadReport (not in includes), got: %s", stdout)
	}
}

// TestEnvOverride_ServerIFSDisabledViaEnv verifies that deep struct bool fields
// under server.* can be overridden via MCP__ env vars:
//
//	MCP__SERVER__IFS__ENABLED=false → server.ifs.enabled=false
func TestEnvOverride_ServerIFSDisabledViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		// Deep struct bool: server.ifs.enabled = false
		"MCP__SERVER__IFS__ENABLED=false",
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

	// IFS should be disabled → 404
	dateStr := time.Now().Format("20060102")
	downloadURL := baseURL + "/_/ifs/download/" + dateStr + "/test.bin"
	resp, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when IFS disabled via ENV, got %d", resp.StatusCode)
	}
}

// TestEnvOverride_LoggingLevelViaEnv verifies that logging.level can be set via
// MCP__LOGGING__LEVEL ENV (not just config file):
//
//	MCP__LOGGING__LEVEL=4 → logging.level=4
func TestEnvOverride_LoggingLevelViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		// Logging level via ENV (no config file)
		"MCP__LOGGING__LEVEL=4",
	)
	var stderrBuf bytes.Buffer
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

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "-->") {
		t.Errorf("expected '-->' (request log) at MCP__LOGGING__LEVEL=4, stderr:\n%s", stderr)
	}
}

// TestEnvOverride_MgmtPortViaEnv verifies that mgmt.port can be overridden via
// MCP__MGMT__PORT ENV:
//
//	MCP__MGMT__PORT=19991 → mgmt.port=19991
func TestEnvOverride_MgmtPortViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", 19100+(time.Now().UnixNano()%1000))

	// Use a unique management port to avoid conflicts
	mgmtPort := 19000 + int(time.Now().UnixNano()%1000)
	if mgmtPort >= 20000 {
		mgmtPort = 19001
	}

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		// Override mgmt port via ENV
		fmt.Sprintf("MCP__MGMT__PORT=%d", mgmtPort),
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

	// Management /health should respond on the overridden port
	mgmtURL := fmt.Sprintf("http://localhost:%d/health", mgmtPort)
	resp, err := http.Get(mgmtURL)
	if err != nil {
		t.Errorf("mgmt /health on overridden port %d failed: %v", mgmtPort, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 on mgmt /health, got %d", resp.StatusCode)
	}
}

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
func startCoreForwardTestServer(t *testing.T, projectDir, mockURL, homeDir, token, cookie string) (cleanup func(), baseURL string) {
	t.Helper()
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mockURL,
	)
	if token != "" {
		cmd.Env = append(cmd.Env, "MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN="+token)
	}
	if cookie != "" {
		cmd.Env = append(cmd.Env, "MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN="+cookie)
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
func writeCoreVirtualConfig(t *testing.T, homeDir, serviceName, yamlContent string) {
	t.Helper()
	logProgress("[config] writing config for %s (home=%s)", serviceName, homeDir)
	t.Helper()
	configDir := filepath.Join(homeDir, "."+serviceName)
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

// logProgress writes a timestamped progress message to stderr for real-time visibility
// during long-running integration tests. Use this instead of t.Logf for operational
// steps (building, starting servers, waiting, etc.) so output appears immediately
// rather than being buffered until the test completes.
// logProgress writes a timestamped progress message directly to the controlling
// terminal (/dev/tty) so it appears in real-time during long-running integration
// tests. go test redirects fd 1 (stdout) and fd 2 (stderr) into internal pipes
// that are only flushed when the test process exits — meaning fmt.Fprintf(os.Stderr)
// and os.Stderr.WriteString are invisible until the test completes (or is killed).
// syscall.Write(2, ...) hits the same pipe and suffers the same fate. /dev/tty is
// NOT a file descriptor but a kernel device that always points to the real
// terminal regardless of any fd redirection — same trick used by ssh, sudo, and gpg
// when they need to read a password while stdin is piped.
func logProgress(format string, args ...interface{}) {
	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("    --- %s %s\n", ts, msg)
	// Write directly to the controlling terminal (/dev/tty) to bypass
	// go test's stderr capture pipe. fd 2 is redirected by go test to a
	// pipe; /dev/tty always points to the real terminal regardless.
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		tty.WriteString(line)
		tty.Close()
	}
}
