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
// Virtual Scenario Mock Upstream Service
// ===========================================================================
// Provides a structured mock upstream with pre-defined response data for
// each virtual pipeline scenario test.

// VirtualMockService provides a mock upstream HTTP server with route-based
// handlers and request recording for virtual tool E2E tests.
type VirtualMockService struct {
	mu       sync.Mutex
	server   *httptest.Server
	requests []AggMockRequest
	routes   map[string]http.HandlerFunc
}

// AggMockRequest records a request received by the mock upstream.
type AggMockRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   []byte
	Header http.Header
}

// NewVirtualMockService creates a new mock service with no routes.
func NewVirtualMockService() *VirtualMockService {
	return &VirtualMockService{
		routes: make(map[string]http.HandlerFunc),
	}
}

// Handle registers a handler for the given path prefix (matched via strings.Contains).
func (m *VirtualMockService) Handle(pathPrefix string, handler http.HandlerFunc) {
	m.routes[pathPrefix] = handler
}

// Start starts the mock HTTP server and returns its base URL.
func (m *VirtualMockService) Start() string {
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Restore body so inner handlers can read it
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		m.mu.Lock()
		m.requests = append(m.requests, AggMockRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Body:   body,
			Header: r.Header.Clone(),
		})
		m.mu.Unlock()

		// Route to matching handler
		for prefix, handler := range m.routes {
			if strings.Contains(r.URL.Path, prefix) {
				handler(w, r)
				return
			}
		}
		// Default: empty JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	return m.server.URL
}

// Close shuts down the mock server.
func (m *VirtualMockService) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// Requests returns a copy of all recorded requests.
func (m *VirtualMockService) Requests() []AggMockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]AggMockRequest, len(m.requests))
	copy(out, m.requests)
	return out
}

// RequestCount returns the number of recorded requests.
func (m *VirtualMockService) RequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// Reset clears recorded requests.
func (m *VirtualMockService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = nil
}

// ===========================================================================
// Pre-defined Scenario Response Data
// ===========================================================================

// RegisterSonatypeIQScenario registers mock handlers that simulate the
// SonatypeIQ API pattern: app info → violations list → component details.
//
// Upstream endpoints (all under /api):
//   GET /api/echo?role=app              → application info
//   GET /api/echo?role=violations&appId= → policy violations list
//   GET /api/hello?name=component-{id}   → component details (per component)
func (m *VirtualMockService) RegisterSonatypeIQScenario() {
	// Component dataset keyed by component ID
	components := map[string]map[string]interface{}{
		"comp-log4j": {
			"componentId":   "comp-log4j",
			"componentName": "log4j-core",
			"version":       "2.14.1",
			"riskScore":     9.8,
			"license":       "Apache-2.0",
		},
		"comp-spring": {
			"componentId":   "comp-spring",
			"componentName": "spring-core",
			"version":       "5.3.9",
			"riskScore":     5.5,
			"license":       "Apache-2.0",
		},
	}

	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		role := r.URL.Query().Get("role")
		switch role {
		case "app":
			writeJSON(w, map[string]interface{}{
				"publicId":        "app-123",
				"name":            "RiskApp",
				"organizationId":  "org-1",
				"internalToken":   "secret-should-be-stripped",
				"debugInfo":       "debug-should-be-stripped",
			})
		case "violations":
			writeJSON(w, []interface{}{
				map[string]interface{}{
					"policyViolationId": "pv-1",
					"severity":          "HIGH",
					"componentId":       "comp-log4j",
					"policyName":        "Critical CVE Policy",
					"waived":            false,
				},
				map[string]interface{}{
					"policyViolationId": "pv-2",
					"severity":          "MEDIUM",
					"componentId":       "comp-spring",
					"policyName":        "License Policy",
					"waived":            true,
				},
			})
		default:
			writeJSON(w, map[string]interface{}{"status": "ok"})
		}
	})

	m.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		// Parse component ID from name: "component-comp-log4j" → "comp-log4j"
		compID := strings.TrimPrefix(name, "component-")
		if comp, ok := components[compID]; ok {
			writeJSON(w, comp)
		} else {
			writeJSON(w, map[string]interface{}{
				"greeting": fmt.Sprintf("Hello, %s!", name),
				"code":     200,
			})
		}
	})
}

// RegisterScenario_ParseJSON registers handlers for testing parse: json on call steps.
// GET /api/echo?format=text → returns JSON as text/plain (requires explicit parse)
func (m *VirtualMockService) RegisterParseJSONScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		if format == "text" {
			// Return JSON as text/plain — requires parse:json on call step
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(`{"nested":{"key":"deep-value","count":42}}`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, map[string]interface{}{"status": "ok"})
		}
	})
}

// RegisterScenario_RequireNonEmpty registers handlers for testing require validation.
// GET /api/echo?role=empty → returns empty list
// GET /api/echo?role=data  → returns non-empty data
func (m *VirtualMockService) RegisterRequireNonEmptyScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		role := r.URL.Query().Get("role")
		switch role {
		case "empty":
			writeJSON(w, map[string]interface{}{"items": []interface{}{}})
		default:
			writeJSON(w, map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			})
		}
	})
}

// RegisterScenario_SimpleChain registers a simple call→jq→return handler.
// GET /api/echo → returns object with fields to project
func (m *VirtualMockService) RegisterSimpleChainScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{
			"status":        "ok",
			"message":       "hello",
			"internalToken": "secret123",
			"timestamp":     "2024-01-01T00:00:00Z",
		})
	})
}

// RegisterScenario_ForeachConcurrency registers handlers for foreach with concurrency.
// GET /api/hello?name=item-{i} → returns a result per item
func (m *VirtualMockService) RegisterForeachConcurrencyScenario() {
	m.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		writeJSON(w, map[string]interface{}{
			"processed": name,
			"status":    "done",
		})
	})
}

// RegisterScenario_ReturnWithVarsExpr registers handlers for return with vars+expr.
// GET /api/echo → returns source data
// GET /api/hello?name=X → returns secondary data
func (m *VirtualMockService) RegisterReturnWithVarsExprScenario() {
	m.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{
			"echo_status": "done",
			"trace_id":    "abc-123",
			"server":      "main",
		})
	})
	m.Handle("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		writeJSON(w, map[string]interface{}{
			"greeting": fmt.Sprintf("Hello, %s!", name),
			"code":     200,
			"ts":       "2024-01-01",
		})
	})
}

// ===========================================================================
// Helpers
// ===========================================================================

// RegisterSonatypeIQRealScenario registers mock handlers that simulate the full
// SonatypeIQ API surface involved in the virtual tool pipeline:
//
//	GET  /api/v2/applications/{appId}/reports/{scanId}/policy  → policy violations
//	GET  /api/v2/reports/applications/{appId}/history           → report history
//	POST /api/v2/components/remediation/{type}/{id}             → remediation suggestions
//
// The response data is realistic, modelled on the actual Sonatype IQ API schemas.
func (m *VirtualMockService) RegisterSonatypeIQRealScenario() {
	// Internal application ID exposed by the policy endpoint
	const internalAppID = "app-internal-uuid-12345"
	const scanID = "b3a29500a885473ca6b4b5c759c39bf2"

	// 1) Policy violations
	// GET /api/v2/applications/{applicationPublicId}/reports/{scanId}/policy
	// Sets .application.id (used later by the history call) and .components[].
	m.Handle("/policy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{
			"application": map[string]interface{}{
				"id":              internalAppID,
				"name":            "bff_peak_transformation_service_fxscenario",
				"organizationId":  "org-abc-def",
				"publicId":        "12609073_bff_peak_transformation_service_fxscenario",
				"contactUserName": "admin",
			},
			"reportTime":  "2024-06-15T08:30:00Z",
			"reportTitle": "Policy Evaluation Report",
			"initiator":   "jenkins",
			"commitHash":  "abc123def456",
			"counts": map[string]interface{}{
				"critical": 1,
				"severe":   1,
				"moderate": 1,
			},
			"components": []interface{}{
				// log4j-core — max threat 9, active violation, remediation exists
				map[string]interface{}{
					"displayName": "log4j-core-2.14.1.jar",
					"hash":        "a1b2c3d4e5f6a1b2",
					"sha256":      "sha256log4jabcdef1234567890abcdef1234567890",
					"packageUrl":  "pkg:maven/org.apache.logging.log4j/log4j-core@2.14.1",
					"originalPurl": "pkg:maven/org.apache.logging.log4j/log4j-core@2.14.1",
					"matchState":  "exact",
					"proprietary": false,
					"thirdParty":  true,
					"componentIdentifier": map[string]interface{}{
						"format": "maven",
						"coordinates": map[string]interface{}{
							"artifactId": "log4j-core",
							"groupId":    "org.apache.logging.log4j",
							"version":    "2.14.1",
						},
					},
					"violations": []interface{}{
						map[string]interface{}{
							"policyViolationId":    "pv-log4j-cve-2021-44228",
							"policyId":             "policy-critical-cve",
							"policyName":           "Critical CVE Policy",
							"policyThreatLevel":    9,
							"policyThreatCategory": "security",
							"waived":               false,
							"grandfathered":        false,
							"waivedWithAutoWaiver": false,
							"constraints": []interface{}{
								map[string]interface{}{
									"constraintId":   "c-44228",
									"constraintName": "CVE-2021-44228",
									"conditions": []interface{}{
										map[string]interface{}{
											"conditionReason":  "CVE-2021-44228",
											"conditionSummary": "Log4Shell remote code execution vulnerability",
										},
									},
								},
							},
						},
					},
				},
				// spring-core — max threat 7, has active + waived violations
				map[string]interface{}{
					"displayName": "spring-core-5.3.9.jar",
					"hash":        "b2c3d4e5f6a2b3c4",
					"sha256":      "sha256springfedcba0987654321fedcba0987654321",
					"packageUrl":  "pkg:maven/org.springframework/spring-core@5.3.9",
					"originalPurl": "pkg:maven/org.springframework/spring-core@5.3.9",
					"matchState":  "exact",
					"proprietary": false,
					"thirdParty":  true,
					"componentIdentifier": map[string]interface{}{
						"format": "maven",
						"coordinates": map[string]interface{}{
							"artifactId": "spring-core",
							"groupId":    "org.springframework",
							"version":    "5.3.9",
						},
					},
					"violations": []interface{}{
						map[string]interface{}{
							"policyViolationId":    "pv-spring-cve-2022-22965",
							"policyId":             "policy-severe-cve",
							"policyName":           "Severe CVE Policy",
							"policyThreatLevel":    7,
							"policyThreatCategory": "security",
							"waived":               false,
							"grandfathered":        false,
							"waivedWithAutoWaiver": false,
							"constraints":          []interface{}{},
						},
						map[string]interface{}{
							"policyViolationId":    "pv-spring-license",
							"policyId":             "policy-license",
							"policyName":           "License Enforcement Policy",
							"policyThreatLevel":    5,
							"policyThreatCategory": "license",
							"waived":               true,
							"grandfathered":        false,
							"waivedWithAutoWaiver": false,
							"constraints":          []interface{}{},
						},
					},
				},
				// safe-lib — max threat 2, below default minThreatLevel of 5
				map[string]interface{}{
					"displayName": "safe-lib-1.0.jar",
					"hash":        "c3d4e5f6a3c4d5e6",
					"sha256":      "sha256safe142536abcdef142536abcdef142536abcdef",
					"packageUrl":  "pkg:maven/com.example/safe-lib@1.0",
					"originalPurl": "pkg:maven/com.example/safe-lib@1.0",
					"matchState":  "exact",
					"proprietary": false,
					"thirdParty":  true,
					"componentIdentifier": map[string]interface{}{
						"format": "maven",
						"coordinates": map[string]interface{}{
							"artifactId": "safe-lib",
							"groupId":    "com.example",
							"version":    "1.0",
						},
					},
					"violations": []interface{}{
						map[string]interface{}{
							"policyViolationId":    "pv-safe-low",
							"policyId":             "policy-low",
							"policyName":           "Low Priority Advisory",
							"policyThreatLevel":    2,
							"policyThreatCategory": "security",
							"waived":               false,
							"grandfathered":        false,
							"waivedWithAutoWaiver": false,
							"constraints":          []interface{}{},
						},
					},
				},
			},
		})
	})

	// 2) Report history
	// GET /api/v2/reports/applications/{applicationId}/history
	m.Handle("/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{
			"applicationId": internalAppID,
			"reports": []interface{}{
				map[string]interface{}{
					"scanId":              "other-scan-001",
					"stage":               "develop",
					"evaluationDate":      "2024-06-14T12:00:00Z",
					"policyEvaluationId":  "eval-001",
					"reportHtmlUrl":       "/report/html/001",
					"reportPdfUrl":        "/report/pdf/001",
					"isReevaluation":      false,
					"isForMonitoring":     false,
					"scanTriggerType":     "WEB_UI",
					"commitHash":          "oldcommit001",
					"affectedComponentCount": 10,
				},
				map[string]interface{}{
					"scanId":              scanID,
					"stage":               "build",
					"evaluationDate":      "2024-06-15T08:30:00Z",
					"policyEvaluationId":  "eval-002",
					"reportHtmlUrl":       "/report/html/002",
					"reportPdfUrl":        "/report/pdf/002",
					"isReevaluation":      true,
					"isForMonitoring":     false,
					"scanTriggerType":     "JENKINS",
					"commitHash":          "abc123def456",
					"affectedComponentCount": 12,
				},
				map[string]interface{}{
					"scanId":              "other-scan-003",
					"stage":               "release",
					"evaluationDate":      "2024-06-15T10:00:00Z",
					"policyEvaluationId":  "eval-003",
					"reportHtmlUrl":       "/report/html/003",
					"reportPdfUrl":        "/report/pdf/003",
					"isReevaluation":      false,
					"isForMonitoring":     true,
					"scanTriggerType":     "CLI",
					"commitHash":          "latestcommit003",
					"affectedComponentCount": 8,
				},
			},
		})
	})

	// 3) Remediation suggestions
	// POST /api/v2/components/remediation/{ownerType}/{ownerId}
	m.Handle("/remediation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Read the request body to identify which component this is
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		var body map[string]interface{}
		json.Unmarshal(bodyBytes, &body)
		displayName, _ := body["displayName"].(string)

		type verChange struct {
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data"`
		}

		var versionChanges []verChange
		switch {
		case strings.Contains(displayName, "log4j"):
			versionChanges = []verChange{
				{
					Type: "next-non-failing",
					Data: map[string]interface{}{
						"component": map[string]interface{}{
							"displayName": "log4j-core-2.17.1.jar",
							"packageUrl":  "pkg:maven/org.apache.logging.log4j/log4j-core@2.17.1",
							"componentIdentifier": map[string]interface{}{
								"coordinates": map[string]interface{}{
									"version": "2.17.1",
								},
							},
						},
					},
				},
				{
					Type: "next-non-failing-with-dependencies",
					Data: map[string]interface{}{
						"component": map[string]interface{}{
							"displayName": "log4j-core-2.18.0.jar",
							"packageUrl":  "pkg:maven/org.apache.logging.log4j/log4j-core@2.18.0",
							"componentIdentifier": map[string]interface{}{
								"coordinates": map[string]interface{}{
									"version": "2.18.0",
								},
							},
						},
					},
				},
			}
		case strings.Contains(displayName, "spring"):
			versionChanges = []verChange{
				{
					Type: "next-non-failing",
					Data: map[string]interface{}{
						"component": map[string]interface{}{
							"displayName": "spring-core-5.3.18.jar",
							"packageUrl":  "pkg:maven/org.springframework/spring-core@5.3.18",
							"componentIdentifier": map[string]interface{}{
								"coordinates": map[string]interface{}{
									"version": "5.3.18",
								},
							},
						},
					},
				},
			}
		default:
			versionChanges = []verChange{}
		}

		writeJSON(w, map[string]interface{}{
			"remediation": map[string]interface{}{
				"versionChanges": versionChanges,
			},
		})
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
