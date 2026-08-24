# Security Policy

## Supported status

Sky Price Estimator is an engineering-beta service. Security fixes are applied to the current `main` line after verification.

## Reporting

Please report suspected vulnerabilities privately to the repository owner rather than opening a public exploit report.

## Current boundaries

The service validates request structure and numeric bounds, caps request-body size, uses explicit server timeouts, runs as a non-root container user, and is scanned with `govulncheck` in CI.

It does **not** provide authentication, tenant isolation, TLS termination, a WAF, durable audit logging, secrets management, or production infrastructure. Deploy it behind an authenticated gateway/reverse proxy when those controls are required.
