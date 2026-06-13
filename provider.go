// Package openfeature is a Rollops feature-flag provider plugin for the
// OpenFeature ecosystem. OpenFeature is an evaluation standard, not a flag
// store, so this plugin targets its reference backend, flagd: it writes a
// flagd flag-configuration document expressing the rollout percentage as
// flagd `fractional` targeting, PUT to a writable flagd sync endpoint. As a
// Rollops rollout steps 10% → 50% → 100%, the fractional split follows.
package openfeature

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.klarlabs.de/rollops/pkg/plugin"
)

// Provider writes flagd flag configuration to a sync endpoint. SyncURL and Token
// come from the plugin's environment (see Config); Environment is supplied per
// call by Rollops and used as a path segment so one endpoint can host several
// environments.
type Provider struct {
	SyncURL string // writable flagd sync endpoint (PUT <SyncURL>/<env>/<flag>)
	Token   string // optional bearer token for the sync endpoint
	HTTP    *http.Client
}

func (p Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

// flagdFlag is a single flagd flag definition.
type flagdFlag struct {
	State          string         `json:"state"`
	DefaultVariant string         `json:"defaultVariant"`
	Variants       map[string]any `json:"variants"`
	Targeting      map[string]any `json:"targeting,omitempty"`
}

// flagdDoc is the flagd flag-configuration document shape.
type flagdDoc struct {
	Flags map[string]flagdFlag `json:"flags"`
}

// ApplyFlag builds a flagd document for the flag — a boolean on/off variant with
// a fractional split at the percentage — and PUTs it to the sync endpoint.
func (p Provider) ApplyFlag(ctx context.Context, c plugin.FlagChange) error {
	if p.SyncURL == "" {
		return fmt.Errorf("openfeature: OPENFEATURE_FLAGD_SYNC_URL is required")
	}
	state := "ENABLED"
	if c.Disabled {
		state = "DISABLED"
	}
	f := flagdFlag{
		State:          state,
		DefaultVariant: "off",
		Variants:       map[string]any{"on": true, "off": false},
		Targeting: map[string]any{
			"fractional": []any{
				[]any{"on", c.Percentage},
				[]any{"off", 100 - c.Percentage},
			},
		},
	}
	doc := flagdDoc{Flags: map[string]flagdFlag{c.Flag: f}}

	u := fmt.Sprintf("%s/%s/%s", strings.TrimRight(p.SyncURL, "/"), c.Environment, c.Flag)
	if err := p.put(ctx, u, doc); err != nil {
		return fmt.Errorf("openfeature: write flagd config for %q: %w", c.Flag, err)
	}
	return nil
}

func (p Provider) put(ctx context.Context, u string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}
	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
