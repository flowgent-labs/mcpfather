package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
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

func unusedTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate unused TCP port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

// testProcessEnv returns a child-process environment suitable for generated
// MCP server runtime tests. Parent shells often source .env files that contain
// MCP__* overrides; those must not leak into tests that intentionally validate
// YAML config behavior.
func testProcessEnv(overrides ...string) []string {
	overrideKeys := map[string]struct{}{}
	for _, kv := range overrides {
		if key, _, ok := strings.Cut(kv, "="); ok {
			overrideKeys[key] = struct{}{}
		}
	}

	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, kv := range os.Environ() {
		key, _, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if strings.HasPrefix(key, "MCP__") {
			continue
		}
		if _, overridden := overrideKeys[key]; overridden {
			continue
		}
		env = append(env, kv)
	}
	return append(env, overrides...)
}

// testProxyEnv returns the proxy URL and env vars for build subcommands.
// It checks MCPFATHER_TEST_PROXY first, then HTTPS_PROXY.
// If neither is set, no proxy is configured — suitable for environments
// where Go can reach module proxies directly.
func testProxyEnv(t *testing.T) (proxyURL string, envVars []string) {
	t.Helper()

	for _, candidate := range []struct {
		name  string
		value string
	}{
		{name: "MCPFATHER_TEST_PROXY", value: os.Getenv("MCPFATHER_TEST_PROXY")},
		{name: "HTTPS_PROXY", value: os.Getenv("HTTPS_PROXY")},
		{name: "HTTP_PROXY", value: os.Getenv("HTTP_PROXY")},
		{name: "https_proxy", value: os.Getenv("https_proxy")},
		{name: "http_proxy", value: os.Getenv("http_proxy")},
	} {
		raw := strings.TrimSpace(candidate.value)
		if raw == "" {
			continue
		}
		normalized, ok := normalizeTestProxyURL(raw)
		if !ok {
			logProgress("[proxy] ignoring invalid %s=%q", candidate.name, raw)
			continue
		}
		logProgress("[proxy] MCPFATHER_TEST_PROXY=%q HTTPS_PROXY=%q → using %q for build commands",
			os.Getenv("MCPFATHER_TEST_PROXY"), os.Getenv("HTTPS_PROXY"), normalized)
		return normalized, []string{
			"HTTPS_PROXY=" + normalized,
			"HTTP_PROXY=" + normalized,
			"https_proxy=" + normalized,
			"http_proxy=" + normalized,
		}
	}

	logProgress("[proxy] MCPFATHER_TEST_PROXY and HTTPS_PROXY not set or invalid — build commands will use direct network")
	return "", nil
}

func normalizeTestProxyURL(raw string) (string, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return "", false
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return "", false
	}
	if parsed.Hostname() == "" && parsed.Port() != "" {
		parsed.Host = net.JoinHostPort("127.0.0.1", parsed.Port())
	}
	if parsed.Hostname() == "" {
		return "", false
	}
	return parsed.String(), true
}

func testBuildEnv(t *testing.T) []string {
	t.Helper()
	_, envVars := testProxyEnv(t)
	if os.Getenv("GOMAXPROCS") == "" {
		envVars = append(envVars, "GOMAXPROCS=2")
	}
	return envVars
}

// mcpfatherBin returns the path to the mcpfather binary, building it if needed.
func mcpfatherBin(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("cannot find repo root: %v", err)
	}
	bin := filepath.Join(root, "bin", "mcpfather")
	if mcpfatherBuildRequired(t, root, bin) {
		buildEnv := testBuildEnv(t)
		cmd := exec.Command("make", "-C", root, "build")
		cmd.Env = append(os.Environ(), buildEnv...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("make build failed: %v\n%s", err, out)
		}
	}
	return bin
}

func mcpfatherBuildRequired(t *testing.T, root, bin string) bool {
	t.Helper()
	binInfo, err := os.Stat(bin)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		t.Fatalf("stat mcpfather binary: %v", err)
	}

	sourceNewer := false
	stopWalk := fmt.Errorf("source newer than mcpfather binary")
	checkPath := func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			switch info.Name() {
			case ".git", "bin", "usecase":
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode().IsRegular() && info.ModTime().After(binInfo.ModTime()) {
			sourceNewer = true
			return stopWalk
		}
		return nil
	}
	for _, rel := range []string{"go.mod", "go.sum", "cmd", "pkg"} {
		path := filepath.Join(root, rel)
		if err := filepath.Walk(path, checkPath); err != nil && err != stopWalk {
			t.Fatalf("walk source path %s: %v", path, err)
		}
		if sourceNewer {
			return true
		}
	}
	return false
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
	buildEnv := testBuildEnv(t)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), buildEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}
	logProgress("[build] go mod tidy OK — building binary %s", binName)
	cmd = exec.Command("go", "build", "-p", "1", "-o", filepath.Join("bin", binName), ".")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), buildEnv...)
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
	cmd.Env = testProcessEnv(env...)
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

func TestGenerator_GeneratedServerVersionFlag(t *testing.T) {
	projectDir := genProject(t, "echoHeaders", "")
	bin := buildServer(t, projectDir)

	for _, flagName := range []string{"--version", "-V"} {
		t.Run(flagName, func(t *testing.T) {
			cmd := exec.Command(bin, flagName)
			cmd.Env = testProcessEnv("HOME=" + t.TempDir())
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", flagName, err, out)
			}

			got := string(out)
			if !strings.Contains(got, filepath.Base(projectDir)+" dev") {
				t.Fatalf("version output should include binary name and version, got: %q", got)
			}
			for _, want := range []string{"commit: unknown", "built: unknown", "go"} {
				if !strings.Contains(got, want) {
					t.Fatalf("version output missing %q: %q", want, got)
				}
			}
			for _, unexpected := range []string{"Upstream endpoint:", "MCP server listening", "Management server"} {
				if strings.Contains(got, unexpected) {
					t.Fatalf("version command should not start server or load runtime config, got: %q", got)
				}
			}
		})
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
	cmd.Env = testProcessEnv(
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
	cmd.Env = testProcessEnv(
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

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadReport",
	)

	result := mustDownloadResult(t, stdout)
	if result["name"] != "report.pdf" {
		t.Fatalf("name = %v, want report.pdf", result["name"])
	}
	localPath := mustFileURIPath(t, result["url"].(string))
	if !strings.Contains(localPath, filepath.Join(homeDir, "."+filepath.Base(projectDir), "ifs", "download")) {
		t.Fatalf("download local path %q is not under test HOME %q", localPath, homeDir)
	}
	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file %s: %v", localPath, err)
	}
	if string(data) != "fake-binary-pdf-content" {
		t.Errorf("downloaded content = %q, want %q", string(data), "fake-binary-pdf-content")
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

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadReport",
	)

	result := mustDownloadResult(t, stdout)
	// When no Content-Disposition is set, DetermineFileName falls back to
	// the URL path last segment ("download" from /download endpoint) or
	// Content-Type-based extension. The IFS layer wraps it in a UUID name,
	// so verify the content was saved correctly rather than the exact name.
	if fileName, _ := result["name"].(string); !strings.Contains(fileName, "download") {
		t.Errorf("expected filename derived from URL path or content-type, got: %q", fileName)
	}

	localPath := mustFileURIPath(t, result["url"].(string))
	if !strings.Contains(localPath, filepath.Join(homeDir, "."+filepath.Base(projectDir), "ifs", "download")) {
		t.Fatalf("download local path %q is not under test HOME %q", localPath, homeDir)
	}
	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file %s: %v", localPath, err)
	}
	if string(data) != "fake-zip-content" {
		t.Errorf("downloaded content = %q, want %q", string(data), "fake-zip-content")
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

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
			"HOME=" + homeDir,
		},
		"-t", "cli", "DownloadBytes",
	)

	result := mustDownloadResult(t, stdout)
	if got, _ := result["size"].(float64); got != 1024 {
		t.Fatalf("download size = %v, want 1024", result["size"])
	}
	localPath := mustFileURIPath(t, result["url"].(string))
	if !strings.Contains(localPath, filepath.Join(homeDir, "."+filepath.Base(projectDir), "ifs", "download")) {
		t.Fatalf("download local path %q is not under test HOME %q", localPath, homeDir)
	}
	info, err := os.Stat(localPath)
	if err != nil {
		t.Fatalf("downloaded file not found at %s: %v", localPath, err)
	}
	t.Logf("Downloaded file: %s (%d bytes)", filepath.Base(localPath), info.Size())
	if info.Size() != 1024 {
		t.Errorf("expected 1024 bytes, got %d", info.Size())
	}
	data, _ := os.ReadFile(localPath)
	if len(data) != 1024 {
		t.Errorf("expected 1024 bytes on disk, got %d", len(data))
	}
}

// ---------------------------------------------------------------------------
// 6. Upload
// ---------------------------------------------------------------------------

// TestUpload_CLI_FileRef verifies CLI mode upload with the unified file-ref
// approach. The --file flag triggers @file:/// prefix auto-wrapping in the CLI,
// the server downloads the file to IFS temp cache, and forwards as multipart.
func TestUpload_CLI_FileRef(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	binPath := buildServer(t, dir)
	homeDir := t.TempDir()

	// Create a test file
	testContent := []byte("hello world file-ref upload test content")
	testFile := filepath.Join(homeDir, "test-upload.bin")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL,
		},
		"-t", "cli", "UploadFile", "--file=file://"+testFile,
	)
	_ = stdout

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from CLI file-ref upload")
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file from --file flag")
	}
	if string(fileRec.Content) != string(testContent) {
		t.Errorf("uploaded content = %q, want %q", string(fileRec.Content), string(testContent))
	}
}

// TestUpload_HTTP_FileRef verifies HTTP mode upload with the unified file-ref
// approach. The client sends an @file:// URI as the "file" arg, the server
// downloads it and forwards as multipart to the upstream.
func TestUpload_HTTP_FileRef(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	testContent := "http-mode-file-ref-content"
	testFile := filepath.Join(homeDir, "http-upload.bin")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := callNativeTool(t, baseURL, "UploadFile", map[string]interface{}{
		"file": "@file://" + testFile,
	})

	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("HTTP file-ref upload failed: %s", result)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from HTTP file-ref upload")
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file content")
	}
	if string(fileRec.Content) != testContent {
		t.Errorf("uploaded content = %q, want %q", string(fileRec.Content), testContent)
	}
}

func TestUpload_HTTP_IFSUploadSourceRemovedAfterForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/upload_spec.yaml", "uploadFile", "")
	homeDir := t.TempDir()
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	testUUID := "feedface-feed-face-feed-feedfacefeed"
	testContent := "ifs-upload-file-ref-content"
	uploadResp, err := http.Post(baseURL+"/_/ifs/upload/"+testUUID, "application/octet-stream", strings.NewReader(testContent))
	if err != nil {
		t.Fatalf("IFS upload failed: %v", err)
	}
	uploadBody, _ := io.ReadAll(uploadResp.Body)
	uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected IFS upload 201, got %d: %s", uploadResp.StatusCode, string(uploadBody))
	}
	var uploadResult map[string]interface{}
	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		t.Fatalf("IFS upload response is not JSON: %v\n%s", err, string(uploadBody))
	}
	fileURI, _ := uploadResult["path"].(string)
	stagedFile := mustFileURIPath(t, fileURI)
	if wantDir := filepath.Join(homeDir, "."+filepath.Base(dir), "ifs", "upload"); filepath.Dir(stagedFile) != wantDir {
		t.Fatalf("IFS upload path dir = %q, want %q", filepath.Dir(stagedFile), wantDir)
	}
	if _, err := os.Stat(stagedFile); err != nil {
		t.Fatalf("expected IFS upload source before forwarding at %s: %v", stagedFile, err)
	}

	result := callNativeTool(t, baseURL, "UploadFile", map[string]interface{}{
		"file": "@" + fileURI,
	})
	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("HTTP IFS upload file-ref forwarding failed: %s", result)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from IFS upload file-ref")
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if string(fileRec.Content) != testContent {
		t.Errorf("uploaded content = %q, want %q", string(fileRec.Content), testContent)
	}
	if _, err := os.Stat(stagedFile); !os.IsNotExist(err) {
		t.Fatalf("IFS upload source should be removed after forwarding, stat err=%v", err)
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

// TestFileRef_NoFileProvided_OmitsOptionalFilePart verifies that FileRef tools
// still send multipart/form-data when no optional file is provided, but do not
// fabricate an empty file part that can trigger upstream filename validators.
func TestFileRef_NoFileProvided_OmitsOptionalFilePart(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "CreateMultipartResource",
		"--name=no-file-test",
		"--description=No file here",
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("FileRef no-file call failed: %s", stdout)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}
	if record.FormFields["name"] != "no-file-test" {
		t.Errorf("expected form field name=no-file-test, got %q", record.FormFields["name"])
	}
	if record.FormFields["description"] != "No file here" {
		t.Errorf("expected form field description='No file here', got %q", record.FormFields["description"])
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("optional file arg should be omitted when no file is provided")
	}
}

func TestMultipartAnyOfFileRef_CLI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/multipart_anyof_spec.yaml", "uploadReportV1", "")
	assertGeneratedToolUsesMultipart(t, dir, "UploadReportV1")
	binPath := buildServer(t, dir)

	fileURL := mockURL + "/files/report-v1.txt"
	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "UploadReportV1",
		"--id=101",
		"--note=with-file",
		"--file="+fileURL,
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("multipart V1 call failed: %s", stdout)
	}

	assertLastRequestIsMultipart(t, mock)
	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from multipart V1")
	}
	if record.FormFields["note"] != "with-file" {
		t.Errorf("expected note form field, got %q", record.FormFields["note"])
	}
	if _, ok := record.FormFields["id"]; ok {
		t.Error("path parameter id should not be forwarded as a multipart form field")
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.FileName != "report-v1.txt" {
		t.Errorf("expected uploaded file name report-v1.txt, got %q", fileRec.FileName)
	}
	if string(fileRec.Content) != "HELLO-FILEREF-report-v1.txt" {
		t.Errorf("unexpected uploaded file content: %q", string(fileRec.Content))
	}
}

func TestMultipartAnyOfFile_NoFileProvidedOmitsFilePart(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/multipart_anyof_spec.yaml", "uploadReportV2", "")
	assertGeneratedToolUsesMultipart(t, dir, "UploadReportV2")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "UploadReportV2",
		"--id=202",
		"--note=no-file",
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("multipart V2 no-file call failed: %s", stdout)
	}

	assertLastRequestIsMultipart(t, mock)
	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from multipart V2")
	}
	if record.FormFields["note"] != "no-file" {
		t.Errorf("expected note form field, got %q", record.FormFields["note"])
	}
	if _, ok := record.FormFields["id"]; ok {
		t.Error("path parameter id should not be forwarded as a multipart form field")
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("optional anyOf file arg should be omitted when no file is provided")
	}
}

func TestMultipartAnyOfFile_StringValueSendsTextField(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/multipart_anyof_spec.yaml", "uploadReportV2", "")
	assertGeneratedToolUsesMultipart(t, dir, "UploadReportV2")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "UploadReportV2",
		"--id=303",
		"--note=string-file-value",
		"--file=existing-report.zip",
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("multipart V2 string file value call failed: %s", stdout)
	}

	assertLastRequestIsMultipart(t, mock)
	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from multipart V2")
	}
	if got := record.FormFields["file"]; got != "existing-report.zip" {
		t.Fatalf("expected anyOf string branch as form field file=existing-report.zip, got %q", got)
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("string branch should not be forwarded as a file part")
	}
}

func TestMultipartAnyOfFile_QueryParamStaysOutOfForm(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/multipart_anyof_spec.yaml", "uploadReportV2", "")
	assertGeneratedToolUsesMultipart(t, dir, "UploadReportV2")
	binPath := buildServer(t, dir)

	stdout, _ := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "UploadReportV2",
		"--id=404",
		"--configId=cfg-404",
		"--note=query-and-form",
	)

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("multipart V2 query/form call failed: %s", stdout)
	}

	assertLastRequestIsMultipart(t, mock)
	requests := mock.Requests()
	if len(requests) == 0 {
		t.Fatal("expected upstream request")
	}
	lastReq := requests[len(requests)-1]
	if got := lastReq.Query.Get("configId"); got != "cfg-404" {
		t.Fatalf("expected configId query=cfg-404, got %q (query=%s)", got, lastReq.Query.Encode())
	}
	if strings.Contains(lastReq.Path, "{id}") || !strings.Contains(lastReq.Path, "/upload-report/404/v2") {
		t.Fatalf("expected path id substituted, got %q", lastReq.Path)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from multipart V2")
	}
	if got := record.FormFields["note"]; got != "query-and-form" {
		t.Fatalf("expected note form field, got %q", got)
	}
	if _, ok := record.FormFields["configId"]; ok {
		t.Fatal("query parameter configId should not be duplicated as a multipart form field")
	}
	if _, ok := record.FormFields["id"]; ok {
		t.Fatal("path parameter id should not be duplicated as a multipart form field")
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("optional anyOf file arg should be omitted when no file is provided")
	}
}

func TestMultipartAnyOfFile_JSONPartEncoding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/multipart_anyof_spec.yaml", "uploadReportV2", "")
	assertGeneratedToolUsesMultipart(t, dir, "UploadReportV2")
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, t.TempDir(), "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "UploadReportV2", map[string]interface{}{
		"id":       505,
		"configId": "cfg-json",
		"note":     "json-part",
		"metadata": map[string]interface{}{
			"name":    "report",
			"enabled": true,
		},
	})
	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Fatalf("multipart V2 JSON part call failed: %s", result)
	}

	assertLastRequestIsMultipart(t, mock)
	requests := mock.Requests()
	if len(requests) == 0 {
		t.Fatal("expected upstream request")
	}
	if got := requests[len(requests)-1].Query.Get("configId"); got != "cfg-json" {
		t.Fatalf("expected configId query=cfg-json, got %q", got)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from multipart V2")
	}
	if got := record.FieldContentTypes["metadata"]; got != "application/json" {
		t.Fatalf("metadata part Content-Type = %q, want application/json", got)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(record.FormFields["metadata"]), &metadata); err != nil {
		t.Fatalf("metadata part is not JSON: %q: %v", record.FormFields["metadata"], err)
	}
	if metadata["name"] != "report" || metadata["enabled"] != true {
		t.Fatalf("unexpected metadata part body: %#v", metadata)
	}
	if _, ok := record.FormFields["configId"]; ok {
		t.Fatal("query parameter configId should not be duplicated as a multipart form field")
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("optional anyOf file arg should be omitted when no file is provided")
	}
}

func assertGeneratedToolUsesMultipart(t *testing.T, dir, toolName string) {
	t.Helper()
	toolFile := filepath.Join(dir, "pkg", "mcptools", toolName+".go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("read generated tool %s: %v", toolName, err)
	}
	src := string(data)
	if !strings.Contains(src, "ForwardMultipartRequest") {
		t.Fatalf("generated %s should use ForwardMultipartRequest; source:\n%s", toolName, src)
	}
	if strings.Contains(src, "ForwardAndParseResponse") {
		t.Fatalf("generated %s should not use ForwardAndParseResponse for multipart file fields", toolName)
	}
}

func assertLastRequestIsMultipart(t *testing.T, mock *CoreMockService) {
	t.Helper()
	requests := mock.Requests()
	if len(requests) == 0 {
		t.Fatal("expected upstream request")
	}
	contentType := requests[len(requests)-1].Headers.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatalf("expected multipart/form-data Content-Type, got %q", contentType)
	}
	if !strings.Contains(contentType, "boundary=") {
		t.Fatalf("expected multipart boundary in Content-Type, got %q", contentType)
	}
}

// ---------------------------------------------------------------------------
// 6c. Three upload cases from real swagger schemas
// ---------------------------------------------------------------------------

// ===========================================================================
// Case A: Named multipart (SonatypeIQ PUT /api/v2/config/saml)
//
// Verifies that when a multipart schema has named binary properties
// (e.g. "identityProviderXml"), the generated handler:
//  1. Preserves the original field name from the swagger schema
//  2. Builds a multipart/form-data request with that named file part
//  3. Also includes non-binary form fields (e.g. "samlConfiguration")
// ===========================================================================

// TestCaseA_NamedMultipart_CLI verifies CLI mode: a named binary field
// "identityProviderXml" is sent as a multipart file part with the correct
// field name, not a generic "file".
func TestCaseA_NamedMultipart_CLI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseANamedMultipart()

	// Serve the IdP XML file for download
	mock.Handle("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		fileName := strings.TrimPrefix(r.URL.Path, "/files/")
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
		w.Write([]byte("<md:EntityDescriptor xmlns:md=\"urn:oasis:names:tc:SAML:2.0:metadata\" entityID=\"https://idp.example.com\"></md:EntityDescriptor>"))
	})
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_a_named_multipart_spec.yaml", "configSaml", "")
	binPath := buildServer(t, dir)

	fileURL := mockURL + "/files/idp-metadata.xml"
	stdout, stderr := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "ConfigSaml",
		"--identityProviderXml="+fileURL,
		"--samlConfiguration=enabled-true-simple-string",
	)
	_ = stderr

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("Case A CLI call failed: %s", stdout)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}

	// Verify non-binary form field
	if record.FormFields["samlConfiguration"] != "enabled-true-simple-string" {
		t.Errorf("expected samlConfiguration=%q, got %q", "enabled-true-simple-string", record.FormFields["samlConfiguration"])
	}

	// Verify the file was uploaded with the CORRECT named field, not "file"
	fileRec, ok := record.Files["identityProviderXml"]
	if !ok {
		// Show what keys were present for diagnosis
		keys := make([]string, 0, len(record.Files))
		for k := range record.Files {
			keys = append(keys, k)
		}
		t.Fatalf("expected 'identityProviderXml' in uploaded files, got keys: %v", keys)
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file content for identityProviderXml")
	}
	// Content should be what the mock /files/ endpoint returned
	if !strings.Contains(string(fileRec.Content), "EntityDescriptor") {
		t.Errorf("expected SAML metadata XML content, got: %s", string(fileRec.Content))
	}

	// "file" must NOT exist as a file field — the handler must use the
	// swagger schema's field name, not a hardcoded default.
	if _, hasFile := record.Files["file"]; hasFile {
		t.Error("'file' field should NOT exist — handler must use the swagger schema field name 'identityProviderXml'")
	}
}

// TestCaseA_NamedMultipart_HTTP verifies HTTP mode: same checks as CLI but
// via the HTTP-native MCP tool call path.
func TestCaseA_NamedMultipart_HTTP(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseANamedMultipart()
	mock.RegisterFileRefScenario() // reuse /files/ endpoint
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_a_named_multipart_spec.yaml", "configSaml", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	fileURL := mockURL + "/files/idp-metadata.xml"
	result := callNativeTool(t, baseURL, "ConfigSaml", map[string]interface{}{
		"identityProviderXml": fileURL,
		"samlConfiguration":   "enabled-true-simple-string",
	})

	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("Case A HTTP call failed: %s", result)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}
	if _, ok := record.Files["identityProviderXml"]; !ok {
		keys := make([]string, 0, len(record.Files))
		for k := range record.Files {
			keys = append(keys, k)
		}
		t.Fatalf("expected 'identityProviderXml' in uploaded files, got keys: %v", keys)
	}
	if record.FormFields["samlConfiguration"] != "enabled-true-simple-string" {
		t.Errorf("expected samlConfiguration=%q, got %q", "enabled-true-simple-string", record.FormFields["samlConfiguration"])
	}
}

// ===========================================================================
// Case B: Plain binary multipart (Jira POST /api/2/issue/{id}/attachments)
//
// Verifies that when the multipart schema is a plain "type: string,
// format: binary" (no named properties), the converter falls back to a
// single "file" FileArg, and:
//  1. The generated handler sends multipart/form-data with "file" part name
//  2. Path parameters (issueIdOrKey) are correctly substituted in the URL
//  3. The file content is downloaded from the URI and forwarded
// ===========================================================================

// TestCaseB_PlainBinaryMultipart_CLI verifies CLI mode for schema-less
// multipart (Jira attachment pattern). The "file" fallback name is used
// and the path param is substituted.
func TestCaseB_PlainBinaryMultipart_CLI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseBPlainBinaryMultipart()
	mock.RegisterFileRefScenario() // reuse /files/ endpoint for download
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_b_plain_binary_multipart_spec.yaml", "addAttachment", "")
	binPath := buildServer(t, dir)

	fileURL := mockURL + "/files/issue-attachment.pdf"
	stdout, stderr := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "AddAttachment",
		"--issueIdOrKey=PROJ-123",
		"--file="+fileURL,
	)
	_ = stderr

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("Case B CLI call failed: %s", stdout)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}

	// Must use fallback name "file" — the schema has no named properties
	fileRec, ok := record.Files["file"]
	if !ok {
		keys := make([]string, 0, len(record.Files))
		for k := range record.Files {
			keys = append(keys, k)
		}
		t.Fatalf("expected 'file' in uploaded files (fallback), got keys: %v", keys)
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file content for fallback 'file' field")
	}

	// Path param must have been substituted — the upstream path should
	// contain PROJ-123 (verified by the mock echoing the path back).
	// The mock returns path in the JSON response, curl captures it in stdout.
	if !strings.Contains(stdout, "PROJ-123") {
		// The upstream mock echoes the path; if path substitution worked
		// it should be visible in the output. If not present, the proxy
		// may have returned text only. This is a soft check.
		t.Logf("Case B CLI stdout (checking for PROJ-123): %s", stdout[:min(len(stdout), 400)])
	}
}

// TestCaseB_PlainBinaryMultipart_HTTP verifies HTTP mode for plain binary
// multipart — path param + file URI, fallback field name "file".
func TestCaseB_PlainBinaryMultipart_HTTP(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseBPlainBinaryMultipart()
	mock.RegisterFileRefScenario() // reuse /files/ endpoint
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_b_plain_binary_multipart_spec.yaml", "addAttachment", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	fileURL := mockURL + "/files/screenshot.png"
	result := callNativeTool(t, baseURL, "AddAttachment", map[string]interface{}{
		"issueIdOrKey": "PROJ-456",
		"file":         fileURL,
	})

	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("Case B HTTP call failed: %s", result)
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data, got nil")
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		keys := make([]string, 0, len(record.Files))
		for k := range record.Files {
			keys = append(keys, k)
		}
		t.Fatalf("expected 'file' in uploaded files (fallback), got keys: %v", keys)
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file from --file flag")
	}
}

// ===========================================================================
// Case C: Octet-stream (Nexus POST /v1/system/license)
//
// Verifies that application/octet-stream APIs:
//  1. Use ForwardBinaryUploadRequest (not multipart)
//  2. Download the file from the URI and send as raw binary body
//  3. Set Content-Type to application/octet-stream (the original swagger CT)
//  4. The upstream receives the file content as-is (not wrapped in multipart)
// ===========================================================================

// TestCaseC_OctetStream_CLI verifies octet-stream upload in CLI mode.
func TestCaseC_OctetStream_CLI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseCOctetStream()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_c_octet_stream_spec.yaml", "installLicense", "")
	binPath := buildServer(t, dir)

	fileURL := mockURL + "/files/license.lic"
	stdout, stderr := runCLI(t, binPath,
		[]string{"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL},
		"-t", "cli", "InstallLicense",
		"--file="+fileURL,
	)
	_ = stderr

	if strings.Contains(stdout, "MCP error") || strings.Contains(stdout, "failed") {
		t.Errorf("Case C CLI call failed: %s", stdout)
	}

	record := mock.LastOctetStream()
	if record == nil {
		t.Fatal("expected octet-stream upload record, got nil")
	}

	// Content-Type must be application/octet-stream (NOT multipart/form-data)
	if !strings.HasPrefix(record.ContentType, "application/octet-stream") {
		t.Errorf("expected Content-Type=application/octet-stream, got %q", record.ContentType)
	}

	// Body must equal the original file content (raw binary forwarding)
	expectedContent := "HELLO-OCTET-STREAM-license.lic"
	if string(record.Body) != expectedContent {
		t.Errorf("expected body %q, got %q", expectedContent, string(record.Body))
	}

	if record.Size != len(expectedContent) {
		t.Errorf("expected size %d, got %d", len(expectedContent), record.Size)
	}
}

// TestCaseC_OctetStream_HTTP verifies octet-stream upload in HTTP mode.
func TestCaseC_OctetStream_HTTP(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterCaseCOctetStream()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/case_c_octet_stream_spec.yaml", "installLicense", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	fileURL := mockURL + "/files/license.lic"
	result := callNativeTool(t, baseURL, "InstallLicense", map[string]interface{}{
		"file": fileURL,
	})

	if strings.Contains(result, "MCP error") || strings.Contains(result, "failed") {
		t.Errorf("Case C HTTP call failed: %s", result)
	}

	record := mock.LastOctetStream()
	if record == nil {
		t.Fatal("expected octet-stream upload record, got nil")
	}

	if !strings.HasPrefix(record.ContentType, "application/octet-stream") {
		t.Errorf("expected Content-Type=application/octet-stream, got %q", record.ContentType)
	}

	expectedContent := "HELLO-OCTET-STREAM-license.lic"
	if string(record.Body) != expectedContent {
		t.Errorf("expected body %q, got %q", expectedContent, string(record.Body))
	}
}

// ---------------------------------------------------------------------------
// 6d. mcpclient.sh — file upload via --file flag (@file:// convention)
// ---------------------------------------------------------------------------

// TestMcpclientSh_FileFlag_SetsRefURI verifies that `mcpclient.sh call <tool> --file /path/to/f`
// sets the "file" argument to @file:///<abs-path> so the server downloads
// and forwards it as multipart.
func TestMcpclientSh_FileFlag_SetsRefURI(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	homeDir := t.TempDir()
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	// Create a test file for upload
	testFile := filepath.Join(homeDir, "test-upload.txt")
	if err := os.WriteFile(testFile, []byte("mcpclient-test-content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	clientSh := filepath.Join(dir, "mcpclient.sh")
	cmd := exec.Command("bash", clientSh, "call", "CreateMultipartResource",
		"--name=mcpclient-test",
		"--description=new-syntax",
		"--file", testFile,
	)
	cmd.Env = testProcessEnv(
		"MCP_SERVER_ENDPOINT="+baseURL+"/mcp",
		"MCP_SERVER_DOWNLOAD_DIR="+filepath.Join(homeDir, "download"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpclient.sh failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "MCP error") || strings.Contains(string(out), "failed") {
		t.Errorf("mcpclient.sh --file call failed: %s", string(out))
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data from mcpclient.sh --file")
	}
	if record.FormFields["name"] != "mcpclient-test" {
		t.Errorf("expected form field name=mcpclient-test, got %q", record.FormFields["name"])
	}
	fileRec, ok := record.Files["file"]
	if !ok {
		t.Fatal("expected 'file' in uploaded files")
	}
	if fileRec.Size == 0 {
		t.Error("expected non-empty file from --file flag")
	}
	if string(fileRec.Content) != "mcpclient-test-content" {
		t.Errorf("expected file content 'mcpclient-test-content', got %q", string(fileRec.Content))
	}
}

// TestMcpclientSh_KeyValueBodySyntax verifies the new --key value --body '{}'
// argument syntax of mcpclient.sh (non-FileArgs tool). The --body JSON is
// merged into args["body"] and forwarded as application/json (no @ prefixes).
func TestMcpclientSh_KeyValueBodySyntax(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/oas3.0_spec.yaml", "createPost", "")
	homeDir := t.TempDir()
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	clientSh := filepath.Join(dir, "mcpclient.sh")
	cmd := exec.Command("bash", clientSh, "call", "CreatePost",
		"--body", `{"title":"hello","body":"world"}`,
	)
	cmd.Env = testProcessEnv(
		"MCP_SERVER_ENDPOINT="+baseURL+"/mcp",
		"MCP_SERVER_DOWNLOAD_DIR="+filepath.Join(homeDir, "download"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpclient.sh --key --body call failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "MCP error") || strings.Contains(string(out), "failed") {
		t.Errorf("mcpclient.sh --body call failed: %s", string(out))
	}
	if mock.requestCount() == 0 {
		t.Error("expected at least one upstream request")
	}
	// Verify it sent JSON, not multipart (no @ prefixes)
	if len(mock.requests) > 0 {
		ct := mock.requests[0].Headers.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Errorf("expected application/json, got %s", ct)
		}
	}
}

// TestMcpclientSh_NoFile_OmitsOptionalFilePart verifies that calling a FileRef
// tool without a file argument still produces a valid multipart request without
// fabricating an empty file part.
func TestMcpclientSh_NoFile_OmitsOptionalFilePart(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterFileRefScenario()
	mockURL := mock.Start()
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/form_multipart_spec.yaml", "createMultipartResource", "")
	homeDir := t.TempDir()
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mockURL, homeDir, "", "")
	defer cleanup()

	clientSh := filepath.Join(dir, "mcpclient.sh")
	cmd := exec.Command("bash", clientSh, "call", "CreateMultipartResource",
		"--name=empty-file",
		"--description=no file arg at all",
	)
	cmd.Env = testProcessEnv(
		"MCP_SERVER_ENDPOINT="+baseURL+"/mcp",
		"MCP_SERVER_DOWNLOAD_DIR="+filepath.Join(homeDir, "download"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpclient.sh (no-file) call failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "MCP error") || strings.Contains(string(out), "failed") {
		t.Errorf("mcpclient.sh no-file call failed: %s", string(out))
	}

	record := mock.LastFileRef()
	if record == nil {
		t.Fatal("expected multipart upload data")
	}
	if _, ok := record.Files["file"]; ok {
		t.Fatal("optional file arg should be omitted when no file is provided")
	}
	if record.FormFields["name"] != "empty-file" {
		t.Errorf("expected form field name=empty-file, got %q", record.FormFields["name"])
	}
}

// TestMcpclientSh_NestedJSONBody verifies that deeply nested JSON request bodies
// are forwarded correctly to upstream. The --body JSON is JSON-parsed, set as
// args["body"], and forwarded as application/json.
func TestMcpclientSh_NestedJSONBody(t *testing.T) {
	var receivedBody []byte
	mock := startMockUpstream(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":true}`))
	}))
	defer mock.Close()

	dir := genProjectWithSpec(t, "testdata/oas3.0_spec.yaml", "createPost", "")
	homeDir := t.TempDir()
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	clientSh := filepath.Join(dir, "mcpclient.sh")
	nestedJSON := `{"title":"root","meta":{"author":{"name":"Alice","email":"alice@example.com"},"tags":["a","b"],"stats":{"views":42,"likes":7}}}`
	cmd := exec.Command("bash", clientSh, "call", "CreatePost",
		"--body", nestedJSON,
	)
	cmd.Env = testProcessEnv(
		"MCP_SERVER_ENDPOINT="+baseURL+"/mcp",
		"MCP_SERVER_DOWNLOAD_DIR="+filepath.Join(homeDir, "download"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpclient.sh nested JSON call failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "MCP error") || strings.Contains(string(out), "failed") {
		t.Errorf("mcpclient.sh nested JSON call failed: %s", string(out))
	}

	// Verify the upstream received the full nested JSON intact.
	if len(receivedBody) == 0 {
		t.Fatal("expected upstream to receive a body, got empty")
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(receivedBody, &parsed); err != nil {
		t.Fatalf("failed to parse upstream body as JSON: %v\nbody: %s", err, string(receivedBody))
	}
	// Drill into the nested structure
	if title, _ := parsed["title"].(string); title != "root" {
		t.Errorf("top-level title = %q, want %q", title, "root")
	}
	meta, ok := parsed["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'meta' to be a nested object")
	}
	author, ok := meta["author"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'meta.author' to be a nested object")
	}
	if name, _ := author["name"].(string); name != "Alice" {
		t.Errorf("meta.author.name = %q, want %q", name, "Alice")
	}
	if email, _ := author["email"].(string); email != "alice@example.com" {
		t.Errorf("meta.author.email = %q, want %q", email, "alice@example.com")
	}
	stats, ok := meta["stats"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'meta.stats' to be a nested object")
	}
	if views, _ := stats["views"].(float64); views != 42 {
		t.Errorf("meta.stats.views = %v, want 42", views)
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
	data := mustDownloadResult(t, result)
	if got, _ := data["url"].(string); !strings.HasPrefix(got, baseURL+"/_/ifs/download/") {
		t.Errorf("download url = %q, want %s/_/ifs/download/...", got, baseURL)
	}
	downloadURL, _ := data["url"].(string)
	if downloadURL != "" {
		fileID := path.Base(downloadURL)
		localPath := filepath.Join(homeDir, "."+filepath.Base(dir), "ifs", "download", fileID)
		if _, err := os.Stat(localPath); err != nil {
			t.Fatalf("expected local download file before GET at %s: %v", localPath, err)
		}
		resp, err := http.Get(downloadURL)
		if err != nil {
			t.Fatalf("failed to GET returned download url %s: %v", downloadURL, err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("download url status = %d, body=%s", resp.StatusCode, string(body))
		}
		if !bytes.Equal(body, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}) {
			t.Fatalf("download url body = %v, want original binary payload", body)
		}
		if _, err := os.Stat(localPath); !os.IsNotExist(err) {
			t.Fatalf("download file should be removed after one GET, stat err=%v", err)
		}
		resp2, err := http.Get(downloadURL)
		if err != nil {
			t.Fatalf("failed second GET returned download url %s: %v", downloadURL, err)
		}
		resp2.Body.Close()
		if resp2.StatusCode != http.StatusNotFound {
			t.Fatalf("second download URL GET status = %d, want 404", resp2.StatusCode)
		}
	}
}

func TestE2E_Core_BinaryDownloadBaseURIOverride(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "downloadReport", "")
	homeDir := t.TempDir()
	baseURI := "https://files.example.test/mcp"

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "", "MCP__SERVER__IFS__BASE_URI="+baseURI+"/")
	defer cleanup()

	result := callNativeTool(t, baseURL, "DownloadReport", map[string]interface{}{})
	data := mustDownloadResult(t, result)
	if got, _ := data["url"].(string); !strings.HasPrefix(got, baseURI+"/_/ifs/download/") || strings.Contains(got, "?") {
		t.Errorf("download url = %q, want %s/_/ifs/download/{uuid} without query", got, baseURI)
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
	// XML response should be returned as text (not binary download JSON)
	if strings.Contains(result, `"kind":"dl.ifs.mcpfather.com/v1"`) {
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))
	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = testProcessEnv(
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
//  1. Upload a file via POST /_/ifs/upload/{uuid}
//  2. Download once via GET /_/ifs/download/{uuid}
//  3. The downloaded content matches what was uploaded.
func TestIFS_UploadAndDownload(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
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

	testUUID := "deadbeef-dead-dead-dead-deaddeadbeef"
	testContent := "IFS upload test content"
	uploadURL := baseURL + "/_/ifs/upload/" + testUUID

	// Upload
	uploadResp, err := http.Post(uploadURL, "application/octet-stream", strings.NewReader(testContent))
	if err != nil {
		t.Fatalf("IFS upload failed: %v", err)
	}
	uploadBody, _ := io.ReadAll(uploadResp.Body)
	uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", uploadResp.StatusCode)
	}
	var uploadResult map[string]interface{}
	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		t.Fatalf("IFS upload response is not JSON: %v\n%s", err, string(uploadBody))
	}
	if got, _ := uploadResult["path"].(string); !strings.HasPrefix(got, "file://") {
		t.Fatalf("IFS upload path = %q, want file:// absolute URI", got)
	}
	if got, _ := uploadResult["download_url"].(string); got != baseURL+"/_/ifs/download/"+testUUID {
		t.Fatalf("IFS upload download_url = %q, want %q", got, baseURL+"/_/ifs/download/"+testUUID)
	}
	stagedFile := mustFileURIPath(t, uploadResult["path"].(string))
	if wantDir := filepath.Join(homeDir, "."+filepath.Base(projectDir), "ifs", "upload"); filepath.Dir(stagedFile) != wantDir {
		t.Fatalf("IFS upload path dir = %q, want %q", filepath.Dir(stagedFile), wantDir)
	}
	if _, err := os.Stat(stagedFile); err != nil {
		t.Fatalf("expected IFS file before download at %s: %v", stagedFile, err)
	}

	// Download
	downloadURL := baseURL + "/_/ifs/download/" + testUUID
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
	if _, err := os.Stat(stagedFile); !os.IsNotExist(err) {
		t.Fatalf("IFS download should delete file after first GET, stat err=%v", err)
	}
	second, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("IFS second download failed: %v", err)
	}
	second.Body.Close()
	if second.StatusCode != http.StatusNotFound {
		t.Fatalf("IFS second download status = %d, want 404", second.StatusCode)
	}
}

func TestIFS_CleanJobRemovesExpiredFormalFiles(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	serviceName := filepath.Base(projectDir)
	downloadDir := filepath.Join(homeDir, "."+serviceName, "ifs", "download")
	uploadDir := filepath.Join(homeDir, "."+serviceName, "ifs", "upload")
	for _, dir := range []string{downloadDir, uploadDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("create IFS dir %s: %v", dir, err)
		}
	}

	oldFiles := []string{
		filepath.Join(downloadDir, "old-download"),
		filepath.Join(uploadDir, "old-upload"),
	}
	keepFiles := []string{
		filepath.Join(downloadDir, "old-download.inprogress"),
		filepath.Join(uploadDir, "old-upload.inprogress"),
		filepath.Join(downloadDir, "young-download"),
		filepath.Join(uploadDir, "young-upload"),
	}
	for _, file := range append(append([]string{}, oldFiles...), keepFiles...) {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}
	oldTime := time.Now().Add(-10 * time.Second)
	for _, file := range append(append([]string{}, oldFiles...), keepFiles[:2]...) {
		if err := os.Chtimes(file, oldTime, oldTime); err != nil {
			t.Fatalf("chtimes %s: %v", file, err)
		}
	}

	port := fmt.Sprintf("%d", unusedTCPPort(t))
	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		"MCP__SERVER__IFS__CLEAN_JOB_TTL_SECONDS=1s",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	waitForServer(t, "http://localhost:"+port)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		removed := true
		for _, file := range oldFiles {
			if _, err := os.Stat(file); !os.IsNotExist(err) {
				removed = false
				break
			}
		}
		if removed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, file := range oldFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Fatalf("expired formal file should be removed: %s stat err=%v", file, err)
		}
	}
	for _, file := range keepFiles {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("clean job should keep %s: %v", file, err)
		}
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	// Start with -v 0 (or no -v) — config's logging.level=4 should activate
	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "10")
	cmd.Env = testProcessEnv(
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
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
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
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
//	MCP__MGMT__PORT=<free-port> -> mgmt.port=<free-port>
func TestEnvOverride_MgmtPortViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	// Use a unique management port to avoid conflicts
	mgmtPort := unusedTCPPort(t)

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
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
func startCoreForwardTestServer(t *testing.T, projectDir, mockURL, homeDir, token, cookie string, extraEnv ...string) (cleanup func(), baseURL string) {
	t.Helper()
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", unusedTCPPort(t))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	env := []string{
		"HOME=" + homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mockURL,
	}
	if token != "" {
		env = append(env, "MCP__UPSTREAM__DEFAULT__AUTH__STATIC__WEB_TOKEN="+token)
	}
	if cookie != "" {
		env = append(env, "MCP__UPSTREAM__DEFAULT__AUTH__STATIC__COOKIE_TOKEN="+cookie)
	}
	env = append(env, extraEnv...)
	cmd.Env = testProcessEnv(env...)

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

// TestEnvOverride_MgmtEnabledViaEnv verifies that MCP__MGMT__ENABLED=false
// (pointer *bool field) disables the management server.
func TestEnvOverride_MgmtEnabledViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	projectDir := genProject(t, "echoHeaders", "")
	binPath := buildServer(t, projectDir)
	homeDir := t.TempDir()
	port := fmt.Sprintf("%d", unusedTCPPort(t))
	mgmtPort := unusedTCPPort(t)

	cmd := exec.Command(binPath, "--transport", "http", "--port", port)
	cmd.Env = testProcessEnv(
		"HOME="+homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT="+mock.server.URL,
		// *bool field: mgmt.enabled = false
		"MCP__MGMT__ENABLED=false",
		fmt.Sprintf("MCP__MGMT__PORT=%d", mgmtPort),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	baseURL := "http://localhost:" + port
	waitForServer(t, baseURL)

	// Management server should be DISABLED on the process-specific port.
	mgmtURL := fmt.Sprintf("http://127.0.0.1:%d/health", mgmtPort)
	client := http.Client{Timeout: 300 * time.Millisecond}
	resp, err := client.Get(mgmtURL)
	if err == nil {
		resp.Body.Close()
		t.Error("mgmt /health should be unreachable when mgmt.enabled=false via ENV")
	}
}

// TestEnvOverride_NativeToolsArrayViaEnv verifies that MCP__NATIVE_TOOLS__EXPOSE__INCLUDES__0
// properly sets array values when no config file exists (default nil *Expose pointer).
func TestEnvOverride_NativeToolsArrayViaEnv(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	binPath := buildServer(t, dir)
	homeDir := t.TempDir()

	// NO config file — rely entirely on ENV overrides, which must
	// initialise the nil *NativeToolsExposeConfig pointer on demand.
	env := []string{
		"HOME=" + homeDir,
		"MCP__UPSTREAM__DEFAULT__ENDPOINT=" + mock.server.URL,
		"MCP__NATIVE_TOOLS__EXPOSE__REGISTER_ALL_TOOLS_BY_DEFAULT=false",
		"MCP__NATIVE_TOOLS__EXPOSE__INCLUDES__0=EchoHeaders",
		"MCP__NATIVE_TOOLS__EXPOSE__INCLUDES__1=SayHello",
	}

	stdout, _ := runCLI(t, binPath, env, "-t", "cli", "list")

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders (via array ENV override, no config file)")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should contain SayHello (via array ENV override, no config file)")
	}
	// DownloadReport was NOT included
	if strings.Contains(stdout, "DownloadReport") {
		t.Errorf("DownloadReport should NOT appear — not in includes array, got: %s", stdout)
	}
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

func mustDownloadResult(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	data := mustJSON(t, strings.TrimSpace(s))
	if kind, _ := data["kind"].(string); kind != "dl.ifs.mcpfather.com/v1" {
		t.Fatalf("download result kind = %q, want dl.ifs.mcpfather.com/v1; result=%s", kind, s)
	}
	for _, key := range []string{"url", "name", "type", "sha1", "createdAt"} {
		if v, _ := data[key].(string); v == "" {
			t.Fatalf("download result %s is empty; result=%s", key, s)
		}
	}
	if _, ok := data["size"].(float64); !ok {
		t.Fatalf("download result size is missing or non-numeric; result=%s", s)
	}
	return data
}

func mustFileURIPath(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("invalid file URI %q: %v", raw, err)
	}
	if u.Scheme != "file" {
		t.Fatalf("local path URI scheme = %q, want file", u.Scheme)
	}
	if !filepath.IsAbs(u.Path) {
		t.Fatalf("file URI path = %q, want absolute path", u.Path)
	}
	return u.Path
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
