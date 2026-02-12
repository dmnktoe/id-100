package sentry

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
)

// Init initializes the Sentry SDK with the provided DSN
func Init(dsn string) error {
	if dsn == "" {
		log.Println("Sentry DSN not configured, skipping Sentry initialization")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}

	log.Println("Sentry initialized successfully")
	return nil
}
