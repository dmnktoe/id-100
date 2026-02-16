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
			// Filter out events in development if needed
			// Or add additional context/filtering here
			return event
		},
	})
	if err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}

	log.Printf("Sentry error tracking initialized (env=%s, release=%s)", opts.Environment, opts.Release)
	return nil
}
