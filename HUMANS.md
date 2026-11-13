# HUMANS.md

Run and use `go-modio`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-modio
```

```go
c := modio.New(modio.Config{Token: token, UserID: 7, Platform: "windows"})
list, err := c.Search(ctx, modio.SearchQuery{GameID: 3659, Sort: modio.SortDownloads})
if err != nil {
	return err
}
for _, m := range list.Data {
	fmt.Println(m.Name, m.Modfile.VersionName(), m.Modfile.MD5())
}
```

## Contracts

| Entry point | Contract |
| --- | --- |
| `New` | Never fails; an empty token makes every call return `ErrNoToken` without sending. |
| `Client` calls | A 401 satisfies `Denied`; a 429 is `*httpkit.RateLimitError`; other statuses are `*httpkit.StatusError`. |
| `Config.UserID` | Sends calls to `UserURL(id)`; `BaseURL` overrides it. |
| `File.MD5` | The API's only digest, lowercase; the caller verifies it. |
| `Dependencies` | Lists only; never installs. |
| `Files`, `Dependencies` | Follow every page up to 100; an answer that claims more is an error. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
