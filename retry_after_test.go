package modio

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rethunk-Tech/go-httpkit"
)

func TestRateLimitRetryAfterZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	c := New(Config{Token: "t", BaseURL: srv.URL})
	_, err := c.Me(context.Background())
	var rl *httpkit.RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("got %v, want RateLimitError", err)
	}
	now := time.Now()
	if rl.Reset.Before(now.Add(55*time.Second)) || rl.Reset.After(now.Add(65*time.Second)) {
		t.Fatalf("Reset %v, want ~60s from now", rl.Reset)
	}
}
