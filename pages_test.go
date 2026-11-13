package modio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func pageHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		n := 100
		if r.URL.Query().Get("_offset") == "100" {
			n = 1
		}
		data := make([]struct{}, n)
		if err := json.NewEncoder(w).Encode(map[string]any{"data": data, "result_total": 101}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}
}

func TestFilesAllPages(t *testing.T) {
	srv := httptest.NewServer(pageHandler(t))
	t.Cleanup(srv.Close)
	files, err := New(Config{Token: "t", BaseURL: srv.URL}).Files(t.Context(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 101 {
		t.Fatalf("got %d files, want 101", len(files))
	}
}

func TestDependenciesAllPages(t *testing.T) {
	srv := httptest.NewServer(pageHandler(t))
	t.Cleanup(srv.Close)
	mods, err := New(Config{Token: "t", BaseURL: srv.URL}).Dependencies(t.Context(), 1, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(mods) != 101 {
		t.Fatalf("got %d mods, want 101", len(mods))
	}
}

func TestPagingStopsAtAServerThatNeverEnds(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": make([]struct{}, MaxPageSize), "result_total": 1 << 40})
	}))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	c := New(Config{Token: "t", BaseURL: srv.URL})
	if _, err := c.Files(ctx, 1, 2); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Files: %v after %d pages", err, hits.Load())
	}
	hits.Store(0)
	if _, err := c.Dependencies(ctx, 1, 2, true); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Dependencies: %v after %d pages", err, hits.Load())
	}
}
