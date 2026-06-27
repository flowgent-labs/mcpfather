package config

import (
	"fmt"
	"os"

	"sonatypeiq-mcp-v1.203.0-01/internal/mcpaggregator/pipeline"
	"gopkg.in/yaml.v3"
)

// Config represents the aggregated tools configuration.
type Config struct {
	AggregateTools []AggregatedToolConfig `yaml:"aggregateTools"`
}

// AggregatedToolConfig defines a single aggregated MCP tool.
type AggregatedToolConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Annotations map[string]interface{} `yaml:"annotations,omitempty"`
	InputSchema map[string]interface{} `yaml:"inputSchema"`
	Pipeline    []pipeline.StepConfig  `yaml:"pipeline"`
}

// LoadConfig reads and parses a config YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// LoadConfigFromHome loads the config from $HOME/.{name}/config.yaml.
func LoadConfigFromHome(name string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Config{}, nil
	}
	return LoadConfig(home + "/." + name + "/config.yaml")
}
