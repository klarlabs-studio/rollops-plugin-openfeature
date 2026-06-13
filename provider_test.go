package openfeature

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.klarlabs.de/rollops/pkg/plugin"
)

func TestApplyFlag_WritesFlagdFractional(t *testing.T) {
	var path, method, auth string
	var doc flagdDoc
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method, auth = r.URL.Path, r.Method, r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &doc)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{SyncURL: srv.URL + "/", Token: "tok", HTTP: srv.Client()}
	err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "checkout", Environment: "production", Percentage: 25})
	if err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if method != http.MethodPut {
		t.Errorf("method = %s, want PUT", method)
	}
	if path != "/production/checkout" {
		t.Errorf("path = %s, want /production/checkout", path)
	}
	if auth != "Bearer tok" {
		t.Errorf("auth = %q, want Bearer tok", auth)
	}
	f, ok := doc.Flags["checkout"]
	if !ok {
		t.Fatalf("flag missing from doc: %+v", doc)
	}
	if f.State != "ENABLED" {
		t.Errorf("state = %s, want ENABLED", f.State)
	}
	frac, _ := f.Targeting["fractional"].([]any)
	on, _ := frac[0].([]any)
	off, _ := frac[1].([]any)
	if on[1].(float64) != 25 || off[1].(float64) != 75 {
		t.Errorf("fractional split = %v/%v, want 25/75", on[1], off[1])
	}
}

func TestApplyFlag_DisabledSetsDisabledState(t *testing.T) {
	var doc flagdDoc
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &doc)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{SyncURL: srv.URL, HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "staging", Disabled: true}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if doc.Flags["f"].State != "DISABLED" {
		t.Errorf("state = %s, want DISABLED", doc.Flags["f"].State)
	}
}

func TestApplyFlag_NoAuthHeaderWithoutToken(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer srv.Close()
	p := Provider{SyncURL: srv.URL, HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p", Percentage: 50}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if auth != "" {
		t.Errorf("auth header set without token: %q", auth)
	}
}

func TestApplyFlag_RequiresSyncURL(t *testing.T) {
	p := Provider{}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p"}); err == nil || !strings.Contains(err.Error(), "SYNC_URL") {
		t.Fatalf("missing sync url must error, got %v", err)
	}
}
