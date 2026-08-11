// Package mcpconfig defines the MCP server configuration types used as
// the canonical source for JSON Schema generation (via reflection). These
// types are auto-copied to generated MCP server projects.
//
// When adding new config fields or sub-structs, only this file and
// config.templ (for LoadConfig/defaultConfig logic) need updating.
// The JSON Schema is derived automatically via reflection.
package mcpconfig

// ---- top-level config ----

// Config is the root configuration for the MCP server.
type Config struct {
	Server       ServerConfig                   `yaml:"server"`
	Upstream     map[string]UpstreamEntryConfig `yaml:"upstream"`
	Tools        NativeToolsConfig              `yaml:"native_tools"`
	Mgmt         MgmtConfig                     `yaml:"mgmt"`
	Logging      LoggingConfig                  `yaml:"logging"`
	VirtualTools []VirtualToolPipelineConfig    `yaml:"virtual_tools"`
}

// ServerConfig holds the inbound-facing (AI agent client MCP request) configuration.
type ServerConfig struct {
	Auth ServerAuthConfig `yaml:"auth"`
	// HTTP server timeouts in seconds. Zero means no timeout (use with
	// caution: connections without timeouts can leak under slow-client attacks).
	// write_timeout_seconds applies to each write; set to 0 for SSE/streaming
	// endpoints where the response is written incrementally over long periods.
	ReadTimeoutSeconds  int `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int `yaml:"write_timeout_seconds"`
	IdleTimeoutSeconds  int `yaml:"idle_timeout_seconds"`
	// Temporary File System (TFS) REST API for binary data plane —
	// provides built-in HTTP endpoints for uploading and downloading binary
	// files, separating the data plane (TFS) from the control plane (JSON-RPC /mcp).
	// Default: enabled. Set to false to disable for pure JSON-RPC deployments.
	TFS TFSConfig `yaml:"tfs"`
}

// TFSConfig controls the Temporary File System (TFS) REST API — built-in
// HTTP endpoints that handle binary file upload/download as a dedicated
// data plane, separate from the JSON-RPC 2.0 control plane at /mcp.
// Upload:  POST /_/tfs/upload/{uuid}
// Download: GET  /_/tfs/download/{uuid}
// Files are stored under ~/.{serviceName}/tfs/{download,upload}/.
type TFSConfig struct {
	Enabled bool `yaml:"enabled"`
	// BaseURI is the public server base URI used to build TFS download URLs
	// returned by tools after binary upstream downloads. Empty means derive it
	// from the inbound /mcp HTTP request Host/X-Forwarded-* headers when
	// available; non-HTTP transports fall back to file:// absolute local URIs.
	BaseURI string `yaml:"base_uri"`
	// CleanJobTTLSeconds controls how long completed TFS temporary files remain
	// on disk before the background clean job removes them. It accepts Go
	// duration strings such as "5m" or an integer number of seconds.
	CleanJobTTLSeconds string `yaml:"clean-job-ttl-seconds"`
}

// UpstreamEntryConfig configures a named upstream backend for http pipeline nodes
// and native tool forwarding. The key "default" is the primary upstream.
type UpstreamEntryConfig struct {
	Endpoint                string             `yaml:"endpoint"`
	EnableMCPSessionForward bool               `yaml:"enable_mcp_session_forward"`
	Auth                    UpstreamAuthConfig `yaml:"auth"`
}

// ---- auth ----
//
// Naming mirrors reverse-proxy terminology (HAProxy/Envoy): "server" is the
// inbound side facing MCP clients (AI agents), "upstream" are the outbound side
// facing upstream APIs. The two are fully decoupled.

// ServerAuthConfig configures inbound OAuth 2.1 bearer token validation.
type ServerAuthConfig struct {
	OIDC ServerOIDCConfig `yaml:"oidc"`
}

// ServerOIDCConfig configures JWT bearer token validation against an OIDC issuer.
type ServerOIDCConfig struct {
	Enabled                           bool     `yaml:"enabled"`
	Issuer                            string   `yaml:"issuer"`
	JWKSURI                           string   `yaml:"jwks_uri"`
	Audience                          string   `yaml:"audience"`
	EnableClientTokenClaimForward     bool     `yaml:"enable_client_token_claim_forward"`
	AdditionalClientTokenClaimForward []string `yaml:"additional_client_token_claim_forward,omitempty"`
}

// UpstreamAuthConfig holds authentication settings for upstream API calls.
type UpstreamAuthConfig struct {
	OIDC   UpstreamOIDCConfig `yaml:"oidc"`
	Static StaticAuthConfig   `yaml:"static"`
}

// UpstreamOIDCConfig configures machine-to-machine OIDC token exchange.
type UpstreamOIDCConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Issuer       string `yaml:"issuer"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	// ClientSecretFile points to a file containing the OIDC client secret.
	// It is primarily used by Kubernetes CSI secret mounts.
	ClientSecretFile string `yaml:"client_secret_file"`
	Scopes           string `yaml:"scopes"`
	GrantType        string `yaml:"grant_type"`
	TokenURL         string `yaml:"token_url"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
}

// StaticAuthConfig holds static credentials for legacy / simple upstream APIs.
type StaticAuthConfig struct {
	WebToken        string `yaml:"web_token"`
	WebTokenFile    string `yaml:"web_token_file"`
	CookieToken     string `yaml:"cookie_token"`
	CookieTokenFile string `yaml:"cookie_token_file"`
}

// ---- tools ----

// NativeToolsConfig controls which native tools are exposed at startup.
type NativeToolsConfig struct {
	Expose *NativeToolsExposeConfig `yaml:"expose"`
}

// NativeToolsExposeConfig defines which native tools are exposed to external AI agents.
type NativeToolsExposeConfig struct {
	RegisterAllToolsByDefault bool     `yaml:"register_all_tools_by_default"`
	Includes                  []string `yaml:"includes"`
	Excludes                  []string `yaml:"excludes"`
}

// ---- mgmt ----

// MgmtConfig controls management endpoints (pprof, metrics, tracing).
type MgmtConfig struct {
	Enabled *bool         `yaml:"enabled"`
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Pprof   PprofConfig   `yaml:"pprof"`
	Otel    OtelConfig    `yaml:"otel"`
	Metrics MetricsConfig `yaml:"metrics"`
}

// IsEnabled returns true for mgmt unless explicitly disabled in config.
func (c MgmtConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// PprofConfig enables Go pprof debug endpoints.
type PprofConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ServerBind string `yaml:"server_bind"`
}

// OtelConfig configures the OpenTelemetry OTLP exporter.
type OtelConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Endpoint   string  `yaml:"endpoint"`
	Protocol   string  `yaml:"protocol"`
	Timeout    int     `yaml:"timeout"`
	SampleRate float64 `yaml:"sample_rate"`
}

// MetricsConfig configures Prometheus metrics export.
type MetricsConfig struct {
	Enabled             *bool                `yaml:"enabled"`
	Prometheus          bool                 `yaml:"prometheus"`
	ExportInterval      string               `yaml:"export_interval"`
	HistogramBoundaries map[string][]float64 `yaml:"histogram_boundaries"`
	Labels              map[string]string    `yaml:"labels"`
}

// IsEnabled returns true for metrics unless explicitly disabled in config.
func (c MetricsConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// ---- logging ----

// LoggingConfig controls server-side request/response logging.
type LoggingConfig struct {
	// Level sets the logging verbosity level (0-10) from the config file,
	// equivalent to the -v CLI flag. Priority: -v CLI > MCP__LOGGING__LEVEL ENV > logging.level in config.yaml.
	// 0=silent, 1-3=status+duration, 4-7=+headers(names), 8-9=+headers(values)+body(len), 10=+body(content)
	Level int `yaml:"level"`
	// AuthVerbose enables logging of Authorization/Cookie header values.
	// Default: false (values are redacted).
	AuthVerbose bool `yaml:"auth_verbose"`
}

// ---- virtual tools ----

// VirtualToolPipelineConfig holds a single virtual tool pipeline definition.
// Virtual tools compose multiple native tools into a single AI-callable tool
// via a declarative pipeline (call -> jq -> foreach -> emit -> return).
// Schema: https://github.com/flowgent-labs/mcpfather/blob/main/.agents/skills/virtual-tool-creator/resources/dsl-schema.json
type VirtualToolPipelineConfig struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	InputSchema map[string]interface{}   `yaml:"input_schema"`
	Pipeline    []map[string]interface{} `yaml:"pipeline"`
	Annotations map[string]interface{}   `yaml:"annotations,omitempty"`
}

// ---- defaults ----

// DefaultConfig returns sensible defaults used when no config file exists.
func DefaultConfig() *Config {
	t := true
	cfg := &Config{
		Server: ServerConfig{
			Auth: ServerAuthConfig{
				OIDC: ServerOIDCConfig{
					EnableClientTokenClaimForward: true,
				},
			},
			ReadTimeoutSeconds:  30,
			WriteTimeoutSeconds: 0, // 0 = disabled (required for SSE streaming)
			IdleTimeoutSeconds:  120,
			TFS: TFSConfig{
				Enabled:            true,
				BaseURI:            "",
				CleanJobTTLSeconds: "5m",
			},
		},
		Upstream: map[string]UpstreamEntryConfig{
			"default": {
				Endpoint:                "https://httpbin.org/anything",
				EnableMCPSessionForward: true,
			},
		},
		Mgmt: MgmtConfig{
			Enabled: &t,
			Host:    "0.0.0.0",
			Port:    9991,
			Metrics: MetricsConfig{
				Enabled:        &t,
				Prometheus:     true,
				ExportInterval: "30s",
				HistogramBoundaries: map[string][]float64{
					"task": {0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
				},
			},
		},
		Logging: LoggingConfig{
			Level:       0,
			AuthVerbose: false,
		},
	}
	return cfg
}
