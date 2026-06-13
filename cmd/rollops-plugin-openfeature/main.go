// Command rollops-plugin-openfeature is a Rollops feature-flag provider plugin
// for the OpenFeature ecosystem, targeting flagd. Build it, pin its sha256, and
// point a rollout's featureFlags.plugin at the binary.
package main

import (
	"fmt"
	"os"

	openfeature "github.com/klarlabs-studio/rollops-plugin-openfeature"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

func main() {
	safety := plugin.Safety{
		// The plugin reaches the flagd sync endpoint, set via
		// OPENFEATURE_FLAGD_SYNC_URL and allow-listed in the host policy.
		NetworkHosts: []string{"flagd.example.com:443"},
		EnvVars:      []string{"OPENFEATURE_FLAGD_SYNC_URL", "OPENFEATURE_TOKEN"},
		RiskClass:    plugin.RiskActive,
	}
	if err := plugin.ServeFlagProvider("klarlabs/openfeature", version, openfeature.FromEnv(), safety); err != nil {
		fmt.Fprintln(os.Stderr, "rollops-plugin-openfeature:", err)
		os.Exit(1)
	}
}
