# Changelog

All notable productization changes are recorded here.

## 0.2.0 - 2026-08-24

- Replaced the generic JSON process stub with deterministic price-estimation logic.
- Added strict request validation, body-size limits, health/readiness endpoints, server timeouts, and request logging.
- Added pricing, validation, HTTP-contract, health, and readiness tests.
- Added `gofmt`, `go vet`, race tests, build, `govulncheck`, Docker, and non-root CI gates.
- Hardened the runtime image to UID/GID 10001.
- Rewrote documentation around the actual engineering-beta product and explicit non-goals.
