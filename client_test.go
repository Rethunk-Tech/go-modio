package modio

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rethunk-Tech/go-httpkit"
)

const timberAPI = `{"id":42,"name":"TimberAPI","summary":"api","profile_url":"https://mod.io/g/timberborn/m/timberapi",
"date_updated":1775230769,"maturity_option":8,"submitted_by":{"username":"eyes"},"tags":[{"name":"Mod"}],
"stats":{"downloads_total":125963,"ratings_positive":143},
"modfile":{"id":9,"filesize":2048,"filename":"t.zip","version":"","filehash":{"md5":"0123456789ABCDEF0123456789abcdef"},"download":{"binary_url":"https://g.modapi.io/f/9"}}}`

func fake(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Modio-Platform") != "windows" {
			t.Errorf("platform header %q", r.Header.Get("X-Modio-Platform"))
		}
		q := r.URL.Query()
		switch r.URL.Path {
		case "/me":
			_, _ = w.Write([]byte(`{"id":7,"username":"nomad"}`))
		case "/games/3659/mods":
			if q.Get("id-in") == "42,43" {
				_, _ = w.Write([]byte(`{"data":[` + timberAPI + `]}`))
				return
			}
			if q.Get("_q") != "timber" || q.Get("_sort") != SortDownloads || q.Get("_limit") != "100" || q.Get("_offset") != "50" ||
				q.Get("tags-in") != "Mod,Tool" || q.Get("tags-not-in") != "Map" {
				t.Errorf("search params %v", q)
			}
			_, _ = w.Write([]byte(`{"data":[` + timberAPI + `],"result_total":3002}`))
		case "/games/3659/mods/42":
			_, _ = w.Write([]byte(timberAPI))
		case "/games/3659/mods/42/files":
			_, _ = w.Write([]byte(`{"data":[{"id":9,"version":"0.8"},{"id":7}]}`))
		case "/games/3659/mods/42/dependencies":
			if q.Get("recursive") != "true" {
				t.Errorf("dependencies params %v", q)
			}
			_, _ = w.Write([]byte(`{"data":[{"id":5,"name":"Harmony"}]}`))
		case "/games/3659/tags":
			_, _ = w.Write([]byte(`{"data":[{"hidden":false,"tags":["Mod","Map"]},{"hidden":true,"tags":["Secret"]}]}`))
		case "/limited":
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return New(Config{Token: " tok ", Platform: "windows", BaseURL: srv.URL})
}

func TestDownloadableNeedsHTTPSAndADigest(t *testing.T) {
	t.Parallel()
	const digest = "0123456789abcdef0123456789abcdef"
	file := func(url, md5 string) File {
		f := File{ID: 9}
		f.Download.BinaryURL = url
		f.Filehash.MD5 = md5
		return f
	}
	if !file("https://g.modapi.io/f/9", digest).Downloadable() {
		t.Error("https with a digest should download")
	}
	if !file("HTTPS://g.modapi.io/f/9", digest).Downloadable() {
		t.Error("uppercase https with a digest should download")
	}
	zero := file("https://g.modapi.io/f/9", digest)
	zero.ID = 0
	for name, f := range map[string]File{
		"http":         file("http://g.modapi.io/f/9", digest),
		"no scheme":    file("g.modapi.io/f/9", digest),
		"no digest":    file("https://g.modapi.io/f/9", ""),
		"short digest": file("https://g.modapi.io/f/9", "abc"),
		"not hex":      file("https://g.modapi.io/f/9", "zz23456789abcdef0123456789abcdef"),
		"no address":   file("", digest),
		"zero id":      zero,
	} {
		if f.Downloadable() {
			t.Errorf("%s: should not download", name)
		}
	}
}

func TestABlankVersionFallsBackToTheFileID(t *testing.T) {
	t.Parallel()
	if got := (File{ID: 9, Version: "  "}).VersionName(); got != "9" {
		t.Fatalf("whitespace version = %q, want 9", got)
	}
}

func TestExplicitIsTheMaturityBit(t *testing.T) {
	t.Parallel()
	if (Mod{}).Explicit() {
		t.Error("zero option should not be explicit")
	}
	if !(Mod{MaturityOption: MaturityExplicit | 1}).Explicit() {
		t.Error("explicit bit among others should be explicit")
	}
}

func TestCalls(t *testing.T) {
	t.Parallel()
	c := fake(t)
	ctx := t.Context()
	list, err := c.Search(ctx, SearchQuery{
		GameID: 3659, Text: "timber", Sort: SortDownloads, Offset: 50,
		TagsIn: []string{"Mod", "Tool"}, TagsNotIn: []string{"Map"},
	})
	if err != nil || list.Total != 3002 || len(list.Data) != 1 {
		t.Fatalf("search = %+v, %v", list, err)
	}
	m := list.Data[0]
	if !m.Explicit() || m.Modfile.VersionName() != "9" || m.Modfile.MD5() != "0123456789abcdef0123456789abcdef" || !m.Modfile.Downloadable() {
		t.Errorf("mod = %+v", m)
	}
	if got, err := c.Search(ctx, SearchQuery{GameID: 3659, IDs: []int{42, 43}}); err != nil || len(got.Data) != 1 {
		t.Errorf("by id = %+v, %v", got, err)
	}
	if got, err := c.Mod(ctx, 3659, 42); err != nil || got.Name != "TimberAPI" {
		t.Errorf("mod = %+v, %v", got, err)
	}
	if files, err := c.Files(ctx, 3659, 42); err != nil || len(files) != 2 || files[0].VersionName() != "0.8" || files[1].VersionName() != "7" {
		t.Errorf("files = %+v, %v", files, err)
	}
	if deps, err := c.Dependencies(ctx, 3659, 42, true); err != nil || len(deps) != 1 || deps[0].Name != "Harmony" {
		t.Errorf("deps = %+v, %v", deps, err)
	}
	if tags, err := c.Tags(ctx, 3659); err != nil || len(tags) != 2 {
		t.Errorf("tags = %v, %v", tags, err)
	}
	if me, err := c.Me(ctx); err != nil || me.ID != 7 || me.Username != "nomad" {
		t.Errorf("me = %+v, %v", me, err)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	c := fake(t)
	bad := New(Config{Token: "nope", BaseURL: c.base})
	if _, err := bad.Me(t.Context()); !Denied(err) {
		t.Errorf("bad token = %v", err)
	}
	if _, err := New(Config{BaseURL: c.base}).Me(t.Context()); !errors.Is(err, ErrNoToken) {
		t.Errorf("no token = %v", err)
	}
	var out any
	if _, ok := httpkitRate(c.get(t.Context(), "/limited", nil, &out)); !ok {
		t.Error("429 is not a rate limit error")
	}
	if UserURL(7) != "https://u-7.modapi.io/v1" || New(Config{Token: "t", UserID: 7}).base != UserURL(7) {
		t.Error("user host")
	}
}

func TestRequestsSendTheConfiguredUserAgent(t *testing.T) {
	t.Parallel()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{Token: "t", BaseURL: srv.URL, UserAgent: "mortar/1"}).Me(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got != "mortar/1" {
		t.Fatalf("User-Agent = %q, want mortar/1", got)
	}
}

func TestAnUnparseableBaseURLIsNotSent(t *testing.T) {
	t.Parallel()
	_, err := New(Config{Token: "t", BaseURL: "http://["}).Me(t.Context())
	if err == nil {
		t.Fatal("unparseable base URL was sent")
	}
}

func TestDependenciesOmitRecursiveWhenNotAsked(t *testing.T) {
	t.Parallel()
	var recursive string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recursive = r.URL.Query().Get("recursive")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{Token: "t", BaseURL: srv.URL}).Dependencies(t.Context(), 1, 2, false); err != nil {
		t.Fatal(err)
	}
	if recursive != "" {
		t.Fatalf("recursive = %q, want omitted", recursive)
	}
}

func httpkitRate(err error) (*httpkit.RateLimitError, bool) {
	rl, ok := errors.AsType[*httpkit.RateLimitError](err)
	return rl, ok && !rl.Reset.IsZero()
}
