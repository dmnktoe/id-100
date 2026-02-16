package sentry

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
)

// InitOptions holds configuration options for Sentry initialization
type InitOptions struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
}

// Init initializes the Sentry SDK with the provided DSN
func Init(dsn string) error {
	return InitWithOptions(InitOptions{
		DSN:              dsn,
		TracesSampleRate: 0.1, // Sample 10% of transactions
	})
}

// InitWithOptions initializes the Sentry SDK with detailed options
func InitWithOptions(opts InitOptions) error {
	if opts.DSN == "" {
		log.Println("Sentry DSN not configured, skipping Sentry initialization")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              opts.DSN,
		Environment:      opts.Environment,
		Release:          opts.Release,
		TracesSampleRate: opts.TracesSampleRate,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Add tags to distinguish backend from frontend
			if event.Tags == nil {
				event.Tags = make(map[string]string)
			}
			event.Tags["layer"] = "backend"
			event.Tags["platform"] = "go"
			return event
		},
	})
	if err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}

	log.Printf("Sentry error tracking initialized (env=%s, release=%s)", opts.Environment, opts.Release)
	return nil
}
