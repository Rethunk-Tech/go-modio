package modio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
	files, err := New(Config{Token: "t", BaseURL: srv.URL}).Files(context.Background(), 1, 2)
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
	mods, err := New(Config{Token: "t", BaseURL: srv.URL}).Dependencies(context.Background(), 1, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(mods) != 101 {
		t.Fatalf("got %d mods, want 101", len(mods))
	}
}
