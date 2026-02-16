/**
 * Sentry error tracking integration for frontend
 */

import * as Sentry from '@sentry/browser';

/**
 * Initialize Sentry error tracking for the browser
 */
export function initSentry(): void {
  const sentryDSN = window.SENTRY_DSN;
  
  if (!sentryDSN) {
    console.log('Sentry DSN not configured, skipping Sentry initialization');
    return;
  }

  try {
    Sentry.init({
      dsn: sentryDSN,
      integrations: [
        Sentry.browserTracingIntegration(),
        Sentry.replayIntegration(),
      ],
      // Performance Monitoring
      tracesSampleRate: 0.1, // Capture 10% of transactions
      // Session Replay
      replaysSessionSampleRate: 0.1, // Sample 10% of sessions
      replaysOnErrorSampleRate: 1.0, // Sample 100% of sessions with errors
      
      environment: window.ENVIRONMENT || 'development',
      release: window.APP_VERSION || 'dev',
      
      beforeSend(event) {
        // Filter out development errors in production
        if (window.ENVIRONMENT === 'development') {
          return null;
        }
        return event;
      },
    });

    console.log('Sentry frontend tracking initialized successfully');
  } catch (error) {
    console.error('Failed to initialize Sentry:', error);
  }
}

/**
 * Capture an exception manually
 */
export function captureException(error: Error | unknown, context?: Record<string, unknown>): void {
  if (context) {
    Sentry.withScope((scope) => {
      Object.entries(context).forEach(([key, value]) => {
        scope.setContext(key, typeof value === 'object' ? value as Record<string, unknown> : { value });
      });
      Sentry.captureException(error);
    });
  } else {
    Sentry.captureException(error);
  }
}

/**
 * Capture a message manually
 */
export function captureMessage(message: string, level: Sentry.SeverityLevel = 'info'): void {
  Sentry.captureMessage(message, level);
}

/**
 * Set user context for Sentry
 */
export function setUser(user: { id?: string; username?: string; email?: string } | null): void {
  Sentry.setUser(user);
}

/**
 * Add breadcrumb for debugging
 */
export function addBreadcrumb(message: string, category?: string, data?: Record<string, unknown>): void {
  Sentry.addBreadcrumb({
    message,
    category: category || 'user-action',
    data,
    level: 'info',
  });
}
