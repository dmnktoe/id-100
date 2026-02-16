/**
 * Utility functions for error handling and common operations
 */

/**
 * Extract error message from unknown error type
 * @param err - Error of unknown type
 * @returns Error message string
 */
export function getErrorMessage(err: unknown): string {
  return err instanceof Error ? err.message : 'Unbekannter Fehler';
}
