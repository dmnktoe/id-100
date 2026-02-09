#!/bin/sh
# Startup script for the webapp container
# This runs before the main application starts

set -e

echo "==> Starting application..."
exec "$@"
