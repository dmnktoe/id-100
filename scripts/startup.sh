#!/bin/sh
# Startup script for the webapp container
# This runs before the main application starts

set -e

# Resolve the app version from the latest GitHub release when it wasn't pinned
# at build time. internal/version picks up APP_VERSION at runtime when the
# binary was built as "dev".
if [ -z "${APP_VERSION}" ] || [ "${APP_VERSION}" = "dev" ]; then
  TAG="$(wget -qO- --header='Accept: application/vnd.github+json' \
    "https://api.github.com/repos/${GITHUB_REPO:-dmnktoe/id-100}/releases/latest" 2>/dev/null \
    | awk -F'"' '{for(i=1;i<=NF;i++) if($i=="tag_name"){print $(i+2); exit}}')"
  [ -n "$TAG" ] && export APP_VERSION="$TAG"
fi

echo "==> Starting application (version=${APP_VERSION:-dev})..."
exec "$@"
