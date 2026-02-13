package middleware

import (
	"testing"

	"id-100/internal/config"
)

func TestInitDatadog(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.DatadogConfig
	}{
		{
			name: "disabled tracer",
			cfg: &config.DatadogConfig{
				Enabled:     false,
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				AgentHost:   "localhost",
				AgentPort:   "8126",
			},
		},
		{
			name: "enabled tracer with defaults",
			cfg: &config.DatadogConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Environment: "test",
				Version:     "1.0.0",
				AgentHost:   "localhost",
				AgentPort:   "8126",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just ensures InitDatadog doesn't panic
			// In a real scenario with an agent running, we could test more thoroughly
			InitDatadog(tt.cfg)
			if tt.cfg.Enabled {
				// Clean up if enabled
				StopDatadog()
			}
		})
	}
}

func TestDatadogMiddleware(t *testing.T) {
	serviceName := "test-service"
	mw := DatadogMiddleware(serviceName)
	
	if mw == nil {
		t.Error("DatadogMiddleware returned nil")
	}
}
