package config

import (
	"os"
	"strconv"
)

// DatadogConfig holds Datadog APM configuration
type DatadogConfig struct {
	Enabled     bool
	ServiceName string
	Environment string
	Version     string
	AgentHost   string
	AgentPort   string
}

// LoadDatadogConfig loads Datadog configuration from environment variables
func LoadDatadogConfig() *DatadogConfig {
	enabled := false
	if ddEnabled := os.Getenv("DD_TRACE_ENABLED"); ddEnabled != "" {
		enabled, _ = strconv.ParseBool(ddEnabled)
	}

	serviceName := os.Getenv("DD_SERVICE")
	if serviceName == "" {
		serviceName = "id-100"
	}

	environment := os.Getenv("DD_ENV")
	if environment == "" {
		environment = os.Getenv("ENVIRONMENT")
		if environment == "" {
			environment = "development"
		}
	}

	version := os.Getenv("DD_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	agentHost := os.Getenv("DD_AGENT_HOST")
	if agentHost == "" {
		agentHost = "localhost"
	}

	agentPort := os.Getenv("DD_TRACE_AGENT_PORT")
	if agentPort == "" {
		agentPort = "8126"
	}

	return &DatadogConfig{
		Enabled:     enabled,
		ServiceName: serviceName,
		Environment: environment,
		Version:     version,
		AgentHost:   agentHost,
		AgentPort:   agentPort,
	}
}
