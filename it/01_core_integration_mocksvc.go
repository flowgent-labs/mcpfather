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
	mu              sync.Mutex
	server          *httptest.Server
	requests        []CoreMockRequest
	routes          map[string]http.HandlerFunc
	lastFileRef     *FileRefRecord
	lastOctetStream *OctetStreamRecord
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

// RegisterUploadScenario registers a POST /upload handler that echoes back
// the uploaded file metadata (content type, size, file content).
func (m *CoreMockService) RegisterUploadScenario() {
	m.Handle("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		bodyBytes, _ := io.ReadAll(r.Body)
		contentType := r.Header.Get("Content-Type")
		var fileContent string
		// Extract file content: try multipart form first, fall back to raw body
		if strings.Contains(contentType, "multipart/form-data") {
			_ = r.ParseMultipartForm(32 << 20)
			if file, _, err := r.FormFile("file"); err == nil {
				data, _ := io.ReadAll(file)
				fileContent = string(data)
				file.Close()
			}
		}
		if fileContent == "" && len(bodyBytes) > 0 {
			fileContent = string(bodyBytes)
		}
		writeCoreJSON(w, map[string]interface{}{
			"method":          r.Method,
			"contentType":     contentType,
			"bodySize":        len(bodyBytes),
			"fileContent":     fileContent,
			"fileContentSize": len(fileContent),
		})
	})
}

// FileRefRecord holds parsed multipart form data recorded by
// RegisterFileRefScenario when a FileRef upload reaches the upstream.
type FileRefRecord struct {
	FormFields        map[string]string
	FieldContentTypes map[string]string
	Files             map[string]FileRefFileRecord
}

// FileRefFileRecord holds metadata about a single uploaded file part.
type FileRefFileRecord struct {
	FileName    string
	ContentType string
	Content     []byte
	Size        int
}

// RegisterFileRefScenario registers:
//   - GET  /files/* → serves a known file body for download
//   - POST /multipart-resource and /upload → parses multipart form, records
//     fields + files, echoes back what was received
//
// Use LastFileRef to inspect multipart data after the call.
func (m *CoreMockService) RegisterFileRefScenario() {
	m.mu.Lock()
	m.lastFileRef = nil
	m.mu.Unlock()

	// Serve files for download
	m.Handle("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Return a predictable payload so tests can verify the uploaded content.
		fileName := strings.TrimPrefix(r.URL.Path, "/files/")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
		w.Write([]byte("HELLO-FILEREF-" + fileName))
	})

	// Shared multipart handler for both paths.
	multipartHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		record := &FileRefRecord{
			FormFields:        make(map[string]string),
			FieldContentTypes: make(map[string]string),
			Files:             make(map[string]FileRefFileRecord),
		}

		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			reader, err := r.MultipartReader()
			if err == nil {
				for {
					part, err := reader.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					name := part.FormName()
					if name == "" {
						part.Close()
						continue
					}
					data, _ := io.ReadAll(part)
					if fileName := part.FileName(); fileName != "" {
						record.Files[name] = FileRefFileRecord{
							FileName:    fileName,
							ContentType: part.Header.Get("Content-Type"),
							Content:     data,
							Size:        len(data),
						}
					} else {
						record.FormFields[name] = string(data)
						record.FieldContentTypes[name] = part.Header.Get("Content-Type")
					}
					part.Close()
				}
			}
		}

		m.mu.Lock()
		m.lastFileRef = record
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"status":     "ok",
			"fieldCount": len(record.FormFields),
			"fileCount":  len(record.Files),
		})
	}

	m.Handle("/multipart-resource", multipartHandler)
	m.Handle("/upload", multipartHandler)
}

// LastFileRef returns the last recorded FileRef multipart upload data.
func (m *CoreMockService) LastFileRef() *FileRefRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastFileRef
}

// ===========================================================================
// Case A: Named multipart (e.g. SonatypeIQ POST /api/v2/config/saml)
// ===========================================================================

// RegisterCaseANamedMultipart registers a handler for /v2/config/saml that
// parses multipart form data and records the field "identityProviderXml"
// (the named binary field) plus the "samlConfiguration" form field.
func (m *CoreMockService) RegisterCaseANamedMultipart() {
	m.mu.Lock()
	m.lastFileRef = nil
	m.mu.Unlock()

	m.Handle("/v2/config/saml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		record := &FileRefRecord{
			FormFields: make(map[string]string),
			Files:      make(map[string]FileRefFileRecord),
		}

		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			_ = r.ParseMultipartForm(32 << 20)
			for name, values := range r.MultipartForm.Value {
				if len(values) > 0 {
					record.FormFields[name] = values[0]
				}
			}
			for name, headers := range r.MultipartForm.File {
				if len(headers) > 0 {
					fh := headers[0]
					f, err := fh.Open()
					if err == nil {
						data, _ := io.ReadAll(f)
						f.Close()
						record.Files[name] = FileRefFileRecord{
							FileName:    fh.Filename,
							ContentType: fh.Header.Get("Content-Type"),
							Content:     data,
							Size:        len(data),
						}
					}
				}
			}
		}

		m.mu.Lock()
		m.lastFileRef = record
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"status":     "ok",
			"fieldCount": len(record.FormFields),
			"fileCount":  len(record.Files),
		})
	})
}

// ===========================================================================
// Case B: Plain binary multipart (e.g. Jira POST /api/2/issue/{id}/attachments)
// ===========================================================================

// RegisterCaseBPlainBinaryMultipart registers a handler for
// /2/issue/*/attachments that parses multipart form data. The upstream
// expects a form part named "file" (the universal fallback for schemas
// without named properties).
func (m *CoreMockService) RegisterCaseBPlainBinaryMultipart() {
	m.mu.Lock()
	m.lastFileRef = nil
	m.mu.Unlock()

	m.Handle("/2/issue/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/attachments") {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		record := &FileRefRecord{
			FormFields: make(map[string]string),
			Files:      make(map[string]FileRefFileRecord),
		}

		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			_ = r.ParseMultipartForm(32 << 20)
			for name, values := range r.MultipartForm.Value {
				if len(values) > 0 {
					record.FormFields[name] = values[0]
				}
			}
			for name, headers := range r.MultipartForm.File {
				if len(headers) > 0 {
					fh := headers[0]
					f, err := fh.Open()
					if err == nil {
						data, _ := io.ReadAll(f)
						f.Close()
						record.Files[name] = FileRefFileRecord{
							FileName:    fh.Filename,
							ContentType: fh.Header.Get("Content-Type"),
							Content:     data,
							Size:        len(data),
						}
					}
				}
			}
		}

		m.mu.Lock()
		m.lastFileRef = record
		m.mu.Unlock()

		// Echo back the issue key that was in the URL path so the test
		// can confirm path-param substitution worked.
		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"status":     "ok",
			"fieldCount": len(record.FormFields),
			"fileCount":  len(record.Files),
			"path":       r.URL.Path,
		})
	})
}

// ===========================================================================
// Case C: Octet-stream (e.g. Nexus POST /v1/system/license)
// ===========================================================================

// OctetStreamRecord holds metadata about an octet-stream upload received by
// RegisterCaseCOctetStream.
type OctetStreamRecord struct {
	ContentType string
	Body        []byte
	Size        int
}

// RegisterCaseCOctetStream registers a handler for /v1/system/license that
// reads the raw request body (application/octet-stream). It also serves a
// /files/* endpoint so the mock can act as both file:// and https:// source
// for the upload test.
func (m *CoreMockService) RegisterCaseCOctetStream() {
	m.mu.Lock()
	m.lastOctetStream = nil
	m.mu.Unlock()

	// File server for download-before-upload
	m.Handle("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		fileName := strings.TrimPrefix(r.URL.Path, "/files/")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
		w.Write([]byte("HELLO-OCTET-STREAM-" + fileName))
	})

	// Octet-stream upload endpoint
	m.Handle("/v1/system/license", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		record := &OctetStreamRecord{
			ContentType: r.Header.Get("Content-Type"),
			Body:        body,
			Size:        len(body),
		}

		m.mu.Lock()
		m.lastOctetStream = record
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		writeCoreJSON(w, map[string]interface{}{
			"status":   "ok",
			"bodySize": record.Size,
		})
	})
}

// LastOctetStream returns the last recorded octet-stream upload data.
func (m *CoreMockService) LastOctetStream() *OctetStreamRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastOctetStream
}

func writeCoreJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
