package modio

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type canned []byte

func (c canned) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(c)), Header: http.Header{}}, nil
}

// FuzzDecoders holds that no answer body makes any call panic.
func FuzzDecoders(f *testing.F) {
	f.Add([]byte(timberAPI))
	f.Add([]byte(`{"data":[` + timberAPI + `],"result_total":1}`))
	f.Add([]byte(`{"data":[{"id":5,"name":"Harmony"},{"tags":[]}]}`))
	f.Add([]byte(`{"data":[{"hidden":false,"tags":["Mod"]}]}`))
	f.Add([]byte(`[`))
	f.Fuzz(func(t *testing.T, body []byte) {
		c := New(Config{Token: "t", BaseURL: "http://modio.test", HTTP: &http.Client{Transport: canned(body)}})
		ctx := t.Context()
		_, _ = c.Search(ctx, SearchQuery{GameID: 1})
		m, _ := c.Mod(ctx, 1, 2)
		_ = m.Explicit()
		files, _ := c.Files(ctx, 1, 2)
		for _, f := range files {
			_ = f.VersionName() + f.MD5()
			_ = f.Downloadable()
		}
		_, _ = c.Dependencies(ctx, 1, 2, true)
		_, _ = c.Tags(ctx, 1)
		_, _ = c.Me(ctx)
	})
}
