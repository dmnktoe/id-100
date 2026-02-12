package config

import (
	"os"
	"testing"
)

func TestLoadDatadogConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected DatadogConfig
	}{
		{
			name: "default values when no env vars set",
			envVars: map[string]string{},
			expected: DatadogConfig{
				Enabled:     false,
				ServiceName: "id-100",
				Environment: "development",
				Version:     "1.0.0",
				AgentHost:   "localhost",
				AgentPort:   "8126",
			},
		},
		{
			name: "custom values from env vars",
			envVars: map[string]string{
				"DD_TRACE_ENABLED":     "true",
				"DD_SERVICE":           "custom-service",
				"DD_ENV":               "production",
				"DD_VERSION":           "2.0.0",
				"DD_AGENT_HOST":        "datadog-agent",
				"DD_TRACE_AGENT_PORT":  "8127",
			},
			expected: DatadogConfig{
				Enabled:     true,
				ServiceName: "custom-service",
				Environment: "production",
				Version:     "2.0.0",
				AgentHost:   "datadog-agent",
				AgentPort:   "8127",
			},
		},
		{
			name: "fallback to ENVIRONMENT for DD_ENV",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
			},
			expected: DatadogConfig{
				Enabled:     false,
				ServiceName: "id-100",
				Environment: "staging",
				Version:     "1.0.0",
				AgentHost:   "localhost",
				AgentPort:   "8126",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg := LoadDatadogConfig()

			// Verify values
			if cfg.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.expected.Enabled)
			}
			if cfg.ServiceName != tt.expected.ServiceName {
				t.Errorf("ServiceName = %v, want %v", cfg.ServiceName, tt.expected.ServiceName)
			}
			if cfg.Environment != tt.expected.Environment {
				t.Errorf("Environment = %v, want %v", cfg.Environment, tt.expected.Environment)
			}
			if cfg.Version != tt.expected.Version {
				t.Errorf("Version = %v, want %v", cfg.Version, tt.expected.Version)
			}
			if cfg.AgentHost != tt.expected.AgentHost {
				t.Errorf("AgentHost = %v, want %v", cfg.AgentHost, tt.expected.AgentHost)
			}
			if cfg.AgentPort != tt.expected.AgentPort {
				t.Errorf("AgentPort = %v, want %v", cfg.AgentPort, tt.expected.AgentPort)
			}
		})
	}
}
