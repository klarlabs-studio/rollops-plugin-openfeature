package openfeature

import "os"

// FromEnv builds a Provider from the plugin's environment. Endpoint and token
// come from the plugin process, never from the Rollops target spec (Rollops
// passes only the flag name, environment, and percentage).
//
//	OPENFEATURE_FLAGD_SYNC_URL  writable flagd sync endpoint (required)
//	OPENFEATURE_TOKEN           optional bearer token for the endpoint
func FromEnv() Provider {
	return Provider{
		SyncURL: os.Getenv("OPENFEATURE_FLAGD_SYNC_URL"),
		Token:   os.Getenv("OPENFEATURE_TOKEN"),
	}
}
