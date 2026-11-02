<h1 align="center">go-modio</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

Client for the mod.io REST API v1: search, sorts, mod details, files, dependencies, per-user API hosts and MD5-checked file addresses. Package `modio`, module `github.com/Rethunk-Tech/go-modio`.

Reads mod.io through a user's own read-only token, with errors a caller can tell apart: bad token, rate limit, other status.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-modio
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- Search a game's mods by text, tags and ids, sorted by popularity, downloads, rating, update or name.
- Mod details, files with MD5 digests, tags and the signed-in account.
- Dependencies, recursive on request.
- Per-user API host (`u-{user-id}.modapi.io`) and the platform header.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.
