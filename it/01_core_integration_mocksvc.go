package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
)

// ===========================================================================
// Core MCP Server Forwarding Scenarios
// ===========================================================================
// Provides a structured mock upstream HTTP server for verifying core MCP
// server request forwarding behaviour: auth, cookies, headers,
// content types, binary responses, and error handling.

// CoreMockService provides a mock upstream HTTP server with request recording.
type CoreMockService struct {
	mu       sync.Mutex
	server   *httptest.Server
	requests []CoreMockRequest
	routes   map[string]http.HandlerFunc
}

// CoreMockRequest records a request received by the mock upstream.
type CoreMockRequest struct {
	Method  string
	Path    string
	Query   url.Values
	Body    []byte
	Headers http.Header
}

// NewCoreMockService creates a new core mock service with no routes.
func NewCoreMockService() *CoreMockService {
	return &CoreMockService{
		routes: make(map[string]http.HandlerFunc),
	}
}

// Handle registers a handler for the given path prefix (matched via strings.Contains).
func (m *CoreMockService) Handle(pathPrefix string, handler http.HandlerFunc) {
	m.routes[pathPrefix] = handler
}

// Start starts the mock HTTP server and returns its base URL.
func (m *CoreMockService) Start() string {
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Restore body so inner handlers can read it
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		m.mu.Lock()
		m.requests = append(m.requests, CoreMockRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.Query(),
			Body:    body,
			Headers: r.Header.Clone(),
		})
		m.mu.Unlock()

		for prefix, handler := range m.routes {
			if strings.Contains(r.URL.Path, prefix) {
				handler(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	return m.server.URL
}

// Close shuts down the mock server.
func (m *CoreMockService) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// Requests returns a copy of all recorded requests.
func (m *CoreMockService) Requests() []CoreMockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]CoreMockRequest, len(m.requests))
	copy(out, m.requests)
	return out
}

// Reset clears recorded requests.
func (m *CoreMockService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = nil
}

// ===========================================================================
// Core Forwarding Scenarios
// ===========================================================================

// RegisterEchoAuthScenario registers an /echo handler that captures and
// echoes back the Authorization header, method, path, and query params.
// Used to verify the MCP server correctly forwards authentication.
func (m *CoreMockService) RegisterEchoAuthScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"authorization": r.Header.Get("Authorization"),
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.Query().Encode(),
			"status":        "ok",
		})
	})
}

// RegisterEchoHeadersScenario registers a handler that echoes all request
// headers (minus host/user-agent). Used to verify header forwarding behaviour.
func (m *CoreMockService) RegisterEchoHeadersScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headers := make(map[string]string)
		for name, values := range r.Header {
			key := strings.ToLower(name)
			if key == "user-agent" || key == "host" {
				continue
			}
			headers[name] = strings.Join(values, "; ")
		}
		writeCoreJSON(w, map[string]interface{}{
			"headers":     headers,
			"contentType": r.Header.Get("Content-Type"),
			"method":      r.Method,
		})
	})
}

// RegisterContentTypeScenario registers handlers that return different
// content types for testing binary/text detection. Uses query param "format"
// on the /echo path (matching EchoHeaders tool) for text types, and handles
// the /download path for binary download tests.
func (m *CoreMockService) RegisterContentTypeScenario() {
	// EchoHeaders tool (GET /api/echo) — dispatch on format query param
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		switch format {
		case "json":
			w.Header().Set("Content-Type", "application/json")
			writeCoreJSON(w, map[string]interface{}{"type": "json", "data": "json-response"})
		case "xml":
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`<root><status>ok</status></root>`))
		case "plain":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("plain text response"))
		default:
			w.Header().Set("Content-Type", "application/json")
			writeCoreJSON(w, map[string]interface{}{"type": "json", "status": "ok"})
		}
	})
	// DownloadReport tool (GET /api/download) — returns binary
	m.Handle("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="data.bin"`)
		w.Write([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05})
	})
}

// RegisterErrorScenario registers handlers on the echo path that return
// error responses based on the "status" query parameter.
func (m *CoreMockService) RegisterErrorScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		statusStr := r.URL.Query().Get("status")
		switch statusStr {
		case "400":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			writeCoreJSON(w, map[string]interface{}{
				"error":   "Bad Request",
				"message": "Invalid parameter",
			})
		case "500":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error"))
		default:
			w.Header().Set("Content-Type", "application/json")
			writeCoreJSON(w, map[string]interface{}{"status": "ok"})
		}
	})
}

// RegisterPathParamScenario registers an /echo handler that echoes back
// query parameters, used to verify path parameter substitution.
func (m *CoreMockService) RegisterPathParamScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"fullPath":  r.URL.Path,
			"fullQuery": r.URL.Query().Encode(),
			"queryName": r.URL.Query().Get("name"),
			"queryAge":  r.URL.Query().Get("age"),
		})
	})
}

// RegisterBodyEchoScenario registers a handler that echoes back the request body.
func (m *CoreMockService) RegisterBodyEchoScenario() {
	m.Handle("/body", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		bodyBytes, _ := io.ReadAll(r.Body)
		var bodyJSON interface{}
		json.Unmarshal(bodyBytes, &bodyJSON)
		writeCoreJSON(w, map[string]interface{}{
			"method":       r.Method,
			"receivedBody": bodyJSON,
			"contentType":  r.Header.Get("Content-Type"),
		})
	})
}

// RegisterGreetingScenario registers a /hello endpoint that returns a greeting.
// Used as the counterpart to the SayHello tool for chained tool tests.
func (m *CoreMockService) RegisterGreetingScenario() {
	m.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		writeCoreJSON(w, map[string]interface{}{
			"greeting": fmt.Sprintf("Hello, %s!", name),
			"code":     200,
		})
	})
}

func writeCoreJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
