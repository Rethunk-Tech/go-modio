# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

- API answers (`Search`, `Mod`, `Files`, `Dependencies`, `Tags`, `Me`) are decoded from JSON capped at 16 MiB and never panic on malformed bodies (`FuzzDecoders`).
- The access token travels in a header only, never in a URL or error text, and a redirect to another origin is refused.

## What is out of scope

The caller verifies a downloaded file against `File.MD5` and checks `File.Download.BinaryURL` before fetching it; the client does not download files.
