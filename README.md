# Sky Price Estimator

**Status: engineering beta.** This repository is a focused Go HTTP service for deterministic price estimation. It is independently deployable and can also sit behind the SKYCOIN4444 API gateway as a pricing utility.

## What it does

`POST /api/v1/estimate` calculates a monetary estimate from unit price, quantity, discount, tax, shipping, and a three-letter currency code. The service validates numeric bounds, rejects unknown JSON fields, limits request bodies, rounds monetary outputs to two decimals, and returns explicit 4xx errors for invalid requests.

Operational endpoints:

- `GET /healthz` — process health
- `GET /readyz` — readiness for traffic

Example request:

```json
{
  "unit_price": 25,
  "quantity": 4,
  "discount_percent": 10,
  "tax_percent": 8.25,
  "shipping": 5,
  "currency": "USD"
}
```

Example result:

```json
{
  "currency": "USD",
  "subtotal": 100,
  "discount": 10,
  "taxable": 90,
  "tax": 7.43,
  "shipping": 5,
  "grand_total": 102.43
}
```

## Run locally

Requires Go 1.23+.

```bash
go test ./...
go run .
```

The service listens on port `8080` by default. Set `PORT` to override it.

## Verification

GitHub Actions gates the product branch with:

- `gofmt` cleanliness
- `go vet`
- race-enabled tests with coverage output
- static build
- `govulncheck`
- Docker image build
- non-root image-user verification

## Container

```bash
docker build -t sky-price .
docker run --rm -p 8080:8080 sky-price
```

The runtime image executes as UID/GID `10001:10001` and contains only the compiled Go service plus the Alpine runtime.

## Architecture

The application intentionally uses only the Go standard library. `estimatePrice` is the domain boundary; HTTP handlers validate transport concerns and call the deterministic calculation function. This keeps the pricing logic reusable from tests or future adapters without coupling it to the HTTP layer.

A SKYCOIN4444 integration should call the HTTP contract or extract the calculation package through a stable versioned interface; it should not copy the service implementation into a flagship repository.

## Security and product boundaries

The service has no authentication or tenant authorization layer and does not persist quotes. TLS is expected to terminate at an ingress or reverse proxy. Currency conversion, dynamic tax jurisdiction rules, catalog pricing, promotions, persistence, audit-grade financial records, HA, and production deployment are **not** implemented or claimed.

See `SECURITY.md` for reporting and operating assumptions.

## License

See `LICENSE`.
