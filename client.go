package modio

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	// BaseURL is the API root for calls that need no per-user host.
	BaseURL = "https://api.mod.io/v1"
	// MaxPageSize is the API's largest _limit.
	MaxPageSize = 100

	requestTimeout = 20 * time.Second
	// maxBody caps one API answer.
	maxBody = 16 << 20
	// maxPages bounds one listing, so an answer that always claims more cannot page forever.
	maxPages = 100
)

// Sort is a value for SearchQuery.Sort; a leading minus sorts descending.
const (
	SortPopular   = "-popular"
	SortDownloads = "-downloads_total"
	SortRating    = "-rating"
	SortUpdated   = "-date_updated"
	SortNewest    = "-date_live"
	SortName      = "name"
)

var errTooManyPages = fmt.Errorf("modio: listing runs past %d pages", maxPages)

// MaturityExplicit is the Mod.MaturityOption bit for explicit content.
const MaturityExplicit = 8

var (
	// ErrNoToken means the Client has no access token; nothing is sent.
	ErrNoToken = errors.New("modio: no access token")
	// ErrNoFile means a mod has no downloadable file.
	ErrNoFile = errors.New("modio: mod has no file to download")
)

// UserURL is the API root mod.io assigns to a user's access token: u-{user-id}.modapi.io.
func UserURL(userID int) string { return "https://u-" + strconv.Itoa(userID) + ".modapi.io/v1" }

// Config is what a Client is built from.
type Config struct {
	// Token is a read-only personal access token, sent as a bearer token. Empty makes every call fail with ErrNoToken.
	Token string
	// UserID, when set, sends calls to the token's own host (UserURL) instead of BaseURL.
	UserID int
	// Platform is sent as X-Modio-Platform ("windows", "linux", ...) and picks which build of a mod's files the API
	// returns; empty sends none.
	Platform string
	// HTTP sends every request; nil means http.DefaultClient. Redirects to another origin are refused so the token
	// stays with the API.
	HTTP *http.Client
	// BaseURL replaces the API root; it wins over UserID.
	BaseURL string
	// UserAgent is sent with every request when set.
	UserAgent string
}

// Client calls the mod.io API. It holds no state beyond its configuration, so it is safe for concurrent use and cheap
// to build again when the token changes.
type Client struct {
	token, base, platform, agent string
	hc                           *http.Client
}

// New builds a Client.
func New(cfg Config) *Client {
	base := BaseURL
	if cfg.UserID > 0 {
		base = UserURL(cfg.UserID)
	}
	return &Client{
		token: strings.TrimSpace(cfg.Token), base: strings.TrimRight(cmp.Or(cfg.BaseURL, base), "/"),
		platform: cfg.Platform, agent: cfg.UserAgent, hc: httpkit.WithRedirect(cfg.HTTP, httpkit.SameOrigin),
	}
}

// Denied reports an answer of 401, which means the token is mistyped, revoked or expired; tokens do not refresh.
func Denied(err error) bool {
	se, ok := errors.AsType[*httpkit.StatusError](err)
	return ok && se.Code == http.StatusUnauthorized
}

// get sends one GET and decodes the 200 answer into out. A 429 is a *httpkit.RateLimitError and any other status a
// *httpkit.StatusError. The token is in a header only, so no error quotes it.
func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	if c.token == "" {
		return ErrNoToken
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("modio: %w", err)
	}
	if c.agent != "" {
		req.Header.Set("User-Agent", c.agent)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	if c.platform != "" {
		req.Header.Set("X-Modio-Platform", c.platform)
	}
	raw, err := httpkit.DoCapped(c.hc, req, maxBody, rateLimited)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("modio: %w", err)
	}
	return nil
}

// rateLimited turns a 429 into a *httpkit.RateLimitError before DoCapped reads its body.
func rateLimited(resp *http.Response) error {
	if resp.StatusCode != http.StatusTooManyRequests {
		return nil
	}
	e := &httpkit.RateLimitError{}
	if t, ok := httpkit.RetryAfter(resp.Header.Get("Retry-After")); ok {
		e.Reset = t
	} else {
		// mod.io documents Retry-After: 0 as wait 60s; httpkit treats 0 as unset
		e.Reset = time.Now().Add(60 * time.Second)
	}
	return e
}

func gamePath(gameID int, rest string) string { return "/games/" + strconv.Itoa(gameID) + rest }

func modPath(gameID, modID int, rest string) string {
	return gamePath(gameID, "/mods/"+strconv.Itoa(modID)+rest)
}

// SearchQuery is one mods listing. Zero fields are not sent.
type SearchQuery struct {
	GameID int
	// Text matches names and summaries.
	Text string
	// Sort is one of the Sort constants; empty keeps the API's order.
	Sort string
	// TagsIn keeps mods carrying any listed tag; TagsNotIn drops mods carrying any.
	TagsIn, TagsNotIn []string
	// IDs keeps only these mods.
	IDs []int
	// Offset is the first mod's position and Limit the page length, at most MaxPageSize; zero Limit is MaxPageSize.
	Offset, Limit int
}

// Search lists a game's live mods.
func (c *Client) Search(ctx context.Context, q SearchQuery) (ModList, error) {
	params := url.Values{}
	if q.Text != "" {
		params.Set("_q", q.Text)
	}
	if q.Sort != "" {
		params.Set("_sort", q.Sort)
	}
	params.Set("_limit", strconv.Itoa(min(cmp.Or(max(q.Limit, 0), MaxPageSize), MaxPageSize)))
	params.Set("_offset", strconv.Itoa(max(q.Offset, 0)))
	if len(q.TagsIn) > 0 {
		params.Set("tags-in", strings.Join(q.TagsIn, ","))
	}
	if len(q.TagsNotIn) > 0 {
		params.Set("tags-not-in", strings.Join(q.TagsNotIn, ","))
	}
	if len(q.IDs) > 0 {
		ids := make([]string, len(q.IDs))
		for i, id := range q.IDs {
			ids[i] = strconv.Itoa(id)
		}
		params.Set("id-in", strings.Join(ids, ","))
	}
	var out ModList
	err := c.get(ctx, gamePath(q.GameID, "/mods"), params, &out)
	return out, err
}

// Mod reads one mod with its current file.
func (c *Client) Mod(ctx context.Context, gameID, modID int) (Mod, error) {
	var out Mod
	err := c.get(ctx, modPath(gameID, modID, ""), nil, &out)
	return out, err
}

// Files lists a mod's files, newest first. A listing past 100 pages is an error.
func (c *Client) Files(ctx context.Context, gameID, modID int) ([]File, error) {
	var all []File
	for offset := 0; offset < maxPages*MaxPageSize; offset += MaxPageSize {
		var out struct {
			Data        []File `json:"data"`
			ResultTotal int    `json:"result_total"`
		}
		params := url.Values{"_sort": {"-date_added"}, "_limit": {strconv.Itoa(MaxPageSize)}, "_offset": {strconv.Itoa(offset)}}
		if err := c.get(ctx, modPath(gameID, modID, "/files"), params, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Data...)
		if len(out.Data) < MaxPageSize || offset+len(out.Data) >= out.ResultTotal {
			return all, nil
		}
	}
	return nil, errTooManyPages
}

// Dependencies lists the mods a mod depends on, all levels down when recursive. Creators also list soft credit links
// here, so a caller offers them and installs none unasked. A listing past 100 pages is an error.
func (c *Client) Dependencies(ctx context.Context, gameID, modID int, recursive bool) ([]Mod, error) {
	var all []Mod
	for offset := 0; offset < maxPages*MaxPageSize; offset += MaxPageSize {
		var out struct {
			Data        []Mod `json:"data"`
			ResultTotal int   `json:"result_total"`
		}
		params := url.Values{"_limit": {strconv.Itoa(MaxPageSize)}, "_offset": {strconv.Itoa(offset)}}
		if recursive {
			params.Set("recursive", "true")
		}
		if err := c.get(ctx, modPath(gameID, modID, "/dependencies"), params, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Data...)
		if len(out.Data) < MaxPageSize || offset+len(out.Data) >= out.ResultTotal {
			return all, nil
		}
	}
	return nil, errTooManyPages
}

// Tags lists the names of a game's visible tags.
func (c *Client) Tags(ctx context.Context, gameID int) ([]string, error) {
	var out struct {
		Data []struct {
			Hidden bool     `json:"hidden"`
			Tags   []string `json:"tags"`
		} `json:"data"`
	}
	if err := c.get(ctx, gamePath(gameID, "/tags"), nil, &out); err != nil {
		return nil, err
	}
	var names []string
	for _, g := range out.Data {
		if !g.Hidden {
			names = append(names, g.Tags...)
		}
	}
	return names, nil
}

// Me reads the token's account.
func (c *Client) Me(ctx context.Context) (User, error) {
	var out User
	err := c.get(ctx, "/me", nil, &out)
	return out, err
}
