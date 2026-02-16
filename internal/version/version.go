package version

import "os"

// Version is set at build time using -ldflags. Default is "dev" for local/local-development builds.
// If the binary was not built with a release tag, we allow a runtime override via the
// APP_VERSION environment variable (useful when Docker was built without --build-arg).
var Version = "dev"

func init() {
	if Version == "dev" {
		if v := os.Getenv("APP_VERSION"); v != "" {
			Version = v
		}
	}
}
