package modio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type answer struct {
	code       int
	retryAfter string
	body       []byte
}

// FuzzHostileAnswer holds that no status, Retry-After or body makes a call panic. One server serves the whole run:
// a server per input runs the machine out of ports.
func FuzzHostileAnswer(f *testing.F) {
	f.Add(200, "", []byte(timberAPI))
	f.Add(200, "", []byte(`{"data":[{}],"result_total":-1}`))
	f.Add(200, "", []byte(strings.Repeat(`{"data":`, 20000)))
	f.Add(429, "99999999999999999999", []byte(nil))
	var a atomic.Pointer[answer]
	a.Store(&answer{code: 200})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := a.Load()
		w.Header()["Retry-After"] = []string{cur.retryAfter}
		w.WriteHeader(cur.code)
		_, _ = w.Write(cur.body)
	}))
	f.Cleanup(srv.Close)
	c := New(Config{Token: "t", BaseURL: srv.URL})
	f.Fuzz(func(t *testing.T, code int, retryAfter string, body []byte) {
		if code < 200 || code > 599 || code/100 == 3 {
			code = 200
		}
		a.Store(&answer{code: code, retryAfter: retryAfter, body: body})
		ctx := t.Context()
		_, _ = c.Search(ctx, SearchQuery{GameID: 1})
		_, _ = c.Mod(ctx, 1, 2)
		_, _ = c.Files(ctx, 1, 2)
		_, _ = c.Dependencies(ctx, 1, 2, true)
		_, _ = c.Tags(ctx, 1)
		_, _ = c.Me(ctx)
	})
}

func TestTheTokenStaysOffARedirect(t *testing.T) {
	t.Parallel()
	var leaked atomic.Value
	other := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		leaked.Store(r.Header.Get("Authorization"))
	}))
	t.Cleanup(other.Close)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/me", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(api.Close)
	if _, err := New(Config{Token: "SECRET", BaseURL: api.URL}).Me(t.Context()); err == nil {
		t.Fatal("a redirect to another origin was followed")
	}
	if v, _ := leaked.Load().(string); v != "" {
		t.Fatalf("the token reached another origin: %q", v)
	}
}

func TestALyingContentLengthFails(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, rw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1099511627776\r\n\r\n{}")
		_ = rw.Flush()
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{Token: "t", BaseURL: srv.URL}).Me(t.Context()); err == nil {
		t.Fatal("a body cut short of its Content-Length was accepted")
	}
}
