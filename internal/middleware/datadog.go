package middleware

import (
	"fmt"
	"log"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	echotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	"github.com/labstack/echo/v4"

	"id-100/internal/config"
)

// InitDatadog initializes the Datadog tracer
func InitDatadog(cfg *config.DatadogConfig) {
	if !cfg.Enabled {
		log.Println("Datadog tracing is disabled")
		return
	}

	opts := []tracer.StartOption{
		tracer.WithService(cfg.ServiceName),
		tracer.WithEnv(cfg.Environment),
		tracer.WithServiceVersion(cfg.Version),
		tracer.WithAgentAddr(fmt.Sprintf("%s:%s", cfg.AgentHost, cfg.AgentPort)),
	}

	tracer.Start(opts...)
	log.Printf("Datadog tracing enabled: service=%s, env=%s, version=%s, agent=%s:%s",
		cfg.ServiceName, cfg.Environment, cfg.Version, cfg.AgentHost, cfg.AgentPort)
}

// StopDatadog stops the Datadog tracer
func StopDatadog() {
	tracer.Stop()
}

// DatadogMiddleware returns Echo middleware for Datadog tracing
func DatadogMiddleware(serviceName string) echo.MiddlewareFunc {
	return echotrace.Middleware(
		echotrace.WithServiceName(serviceName),
	)
}
