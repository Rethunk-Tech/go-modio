# Changelog

Notable changes to `go-modio`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-11-07

### Changed

- `File.Downloadable` is false for an `http://` address and for a file with no well-formed MD5, so a caller can't download in cleartext or install a file it can't verify.

## [0.1.0] - 2026-11-02

### Added

- `Client`: Search (text, tags, ids, sorts), Mod, Files with MD5, Dependencies (recursive), Tags and Me over the mod.io API with a read-only token, the per-user host and the platform header.
- `Denied` and `*httpkit.RateLimitError` to tell a bad token from a rate limit.

[Unreleased]: https://github.com/Rethunk-Tech/go-modio/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/Rethunk-Tech/go-modio/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Rethunk-Tech/go-modio/releases/tag/v0.1.0
