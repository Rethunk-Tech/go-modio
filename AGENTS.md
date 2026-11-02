# AGENTS.md

`modio` is a read-only client for the mod.io REST API v1, signed with a user's personal access token. Mortar imports it; it imports only go-httpkit and never a consumer package.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH.

## Shape

- `Client.get` is the single request path: it refuses an empty token before sending (ErrNoToken), caps the body, refuses cross-origin redirects so the token stays with the API, and turns a 429 into a `*httpkit.RateLimitError` carrying Retry-After.
- `File.Download.BinaryURL` can expire, so a caller reads the file again right before each download; the client never caches one.
- The only hash the API gives is MD5; `File.MD5` lowercases it, and a caller checks it after downloading.
- Dependencies include soft credit links, so `Dependencies` only lists; installing is the caller's decision.

## Files

| File | Holds |
| --- | --- |
| `client.go` | Config, Client, the request path, Search, Mod, Files, Dependencies, Tags, Me, Denied, UserURL |
| `types.go` | Mod, File, ModList, User and their helpers |
