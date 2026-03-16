package version

// AppVersion is the current application version.
// It is overridden at build time via:
//
//	go build -ldflags "-X github.com/mitchellbauer/data-coupler/internal/version.AppVersion=X.Y.Z"
var AppVersion = "0.4.0"
