package mcpconfig

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Top-level keys
	if cfg.Server.Auth.OIDC.EnableClientTokenClaimForward != true {
		t.Error("server.auth.oidc.enable_client_token_claim_forward should default to true")
	}
	if cfg.Server.ReadTimeoutSeconds != 30 {
		t.Errorf("server.read_timeout_seconds = %d, want 30", cfg.Server.ReadTimeoutSeconds)
	}
	if cfg.Server.WriteTimeoutSeconds != 0 {
		t.Errorf("server.write_timeout_seconds = %d, want 0 (disabled for SSE)", cfg.Server.WriteTimeoutSeconds)
	}
	if cfg.Server.IdleTimeoutSeconds != 120 {
		t.Errorf("server.idle_timeout_seconds = %d, want 120", cfg.Server.IdleTimeoutSeconds)
	}

	defaultUp, ok := cfg.Upstream["default"]
	if !ok {
		t.Fatal(`upstream["default"] not found in DefaultConfig`)
	}
	if defaultUp.Endpoint != "https://httpbin.org/anything" {
		t.Errorf(`upstream["default"].endpoint = %q`, defaultUp.Endpoint)
	}
	if !defaultUp.EnableMCPSessionForward {
		t.Error(`upstream["default"].enable_mcp_session_forward should default to true`)
	}

	if !cfg.Mgmt.IsEnabled() {
		t.Error("mgmt should be enabled by default")
	}
	if !cfg.Mgmt.Metrics.IsEnabled() {
		t.Error("mgmt.metrics should be enabled by default")
	}
}

func TestDefaultConfig_YAMLTags(t *testing.T) {
	cfg := DefaultConfig()

	// The Config struct defines the canonical YAML structure used by print-default-config.
	// Verify the config values are accessible through the correct accessor chain.
	if cfg.Upstream["default"].Endpoint == "" {
		t.Error("default upstream endpoint is empty")
	}
	if cfg.Server.ReadTimeoutSeconds <= 0 {
		t.Error("server.read_timeout_seconds should be positive")
	}
}

func TestMgmtConfig_IsEnabled(t *testing.T) {
	t.Run("nil enabled means true", func(t *testing.T) {
		m := MgmtConfig{}
		if !m.IsEnabled() {
			t.Error("nil Enabled should default to true")
		}
	})
	t.Run("explicit false", func(t *testing.T) {
		f := false
		m := MgmtConfig{Enabled: &f}
		if m.IsEnabled() {
			t.Error("explicit false should be false")
		}
	})
	t.Run("explicit true", func(t *testing.T) {
		tr := true
		m := MgmtConfig{Enabled: &tr}
		if !m.IsEnabled() {
			t.Error("explicit true should be true")
		}
	})
}

func TestMetricsConfig_IsEnabled(t *testing.T) {
	t.Run("nil enabled means true", func(t *testing.T) {
		m := MetricsConfig{}
		if !m.IsEnabled() {
			t.Error("nil Enabled should default to true")
		}
	})
	t.Run("explicit false", func(t *testing.T) {
		f := false
		m := MetricsConfig{Enabled: &f}
		if m.IsEnabled() {
			t.Error("explicit false should be false")
		}
	})
}

func TestUpstreamEntryConfig_Fields(t *testing.T) {
	entry := UpstreamEntryConfig{
		Endpoint:                "https://api.example.com",
		EnableMCPSessionForward: true,
		Auth: UpstreamAuthConfig{
			OIDC: UpstreamOIDCConfig{
				Enabled:      true,
				Issuer:       "https://idp.example.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Scopes:       "openid",
			},
			Static: StaticAuthConfig{
				WebToken:     "token123",
				WebTokenFile: "/path/to/token",
			},
		},
	}
	if entry.Endpoint != "https://api.example.com" {
		t.Error("endpoint mismatch")
	}
	if !entry.Auth.OIDC.Enabled {
		t.Error("oidc should be enabled")
	}
	if entry.Auth.Static.WebToken != "token123" {
		t.Error("static web_token mismatch")
	}
}

func TestIFSConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Server.IFS.Enabled {
		t.Error("server.ifs.enabled should default to true")
	}
	if cfg.Server.IFS.BaseURI != "" {
		t.Errorf("server.ifs.base_uri = %q, want empty default", cfg.Server.IFS.BaseURI)
	}
	if cfg.Server.IFS.CleanJobTTLSeconds != "5m" {
		t.Errorf("server.ifs.clean-job-ttl-seconds = %q, want 5m", cfg.Server.IFS.CleanJobTTLSeconds)
	}
}

func TestLoggingConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Logging.Level != 0 {
		t.Errorf("logging.level = %d, want 0", cfg.Logging.Level)
	}
	if cfg.Logging.AuthVerbose != false {
		t.Error("logging.auth_verbose should default to false")
	}
}

func TestLoggingConfig_Level(t *testing.T) {
	l := LoggingConfig{Level: 5, AuthVerbose: true}
	if l.Level != 5 {
		t.Errorf("Level = %d, want 5", l.Level)
	}
	if !l.AuthVerbose {
		t.Error("AuthVerbose should be true")
	}
}

func TestVirtualToolPipelineConfig_Fields(t *testing.T) {
	vt := VirtualToolPipelineConfig{
		Name:        "my-virtual-tool",
		Description: "A composed virtual tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{"type": "string"},
			},
		},
		Pipeline: []map[string]interface{}{
			{"call": map[string]interface{}{"tool": "nativeTool"}},
		},
		Annotations: map[string]interface{}{
			"readOnlyHint": true,
		},
	}
	if vt.Name != "my-virtual-tool" {
		t.Errorf("Name = %q, want 'my-virtual-tool'", vt.Name)
	}
	if vt.Description != "A composed virtual tool" {
		t.Error("Description mismatch")
	}
	if len(vt.Pipeline) != 1 {
		t.Errorf("Pipeline length = %d, want 1", len(vt.Pipeline))
	}
}

func TestIFSConfig_Fields(t *testing.T) {
	ifs := IFSConfig{Enabled: true, BaseURI: "https://files.example.com/mcp", CleanJobTTLSeconds: "30s"}
	if !ifs.Enabled {
		t.Error("Enabled should be true")
	}
	if ifs.BaseURI != "https://files.example.com/mcp" {
		t.Error("BaseURI mismatch")
	}
	if ifs.CleanJobTTLSeconds != "30s" {
		t.Error("CleanJobTTLSeconds mismatch")
	}
	ifs2 := IFSConfig{Enabled: false}
	if ifs2.Enabled {
		t.Error("Enabled should be false")
	}
}
