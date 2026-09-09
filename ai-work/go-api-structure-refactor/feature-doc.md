# Go API structure refactor

The Go API now separates transport, application orchestration, pure QR
calculation, configuration, and the outbound statistics adapter:

```text
handler (Fiber/JSON/CORS) -> usecase.Service -> qr.Factorize
                                             -> external/statistics.Client -> Node API
```

`main.go` is the composition root. It loads configuration, creates one shared
`net/http.Client`, wires the statistics client and use case, and starts Fiber.
The refactor preserves the existing API behavior and does not add dependencies.

## Client-facing endpoint

`POST /qr` accepts a JSON body containing a rectangular numeric matrix:

```json
{"matrix":[[1,2],[3,4]]}
```

On success it returns the QR matrices and the statistics obtained from Node:

```json
{
  "q": [[-0.316227766,-0.948683298],[-0.948683298,0.316227766]],
  "r": [[-3.16227766,-4.427188724],[0, -1.897366596]],
  "statistics": {
    "maximum": 0,
    "minimum": 0,
    "sum": 0,
    "average": 0,
    "hasDiagonalMatrix": false
  }
}
```

The numerical values above are illustrative; clients should treat the returned
QR and statistics values as data rather than relying on a particular sign
convention. Errors retain the existing shape `{ "error", "message" }`:

- `400 invalid_request` for malformed JSON or invalid matrices;
- `502 downstream_error` when the statistics service is unavailable or returns
  an invalid response;
- `504 downstream_timeout` when the statistics request times out;
- `500 internal_error` for unexpected failures.

CORS and `OPTIONS` preflight behavior remain configured by the Fiber handler.

## Node statistics contract

The outbound adapter posts to `POST /api/v1/statistics` at `NODE_API_URL` with:

```json
{"q": [[/* QR matrix */]], "r": [[/* R matrix */]]}
```

It requires the JSON response fields `maximum`, `minimum`, `sum`,
`average`, and `hasDiagonalMatrix`. Numeric fields must be finite. The adapter
uses the request context, closes response bodies, rejects non-2xx responses,
and classifies timeout and other downstream failures for the HTTP handler.

## Runtime configuration

Configuration is read from environment variables (see `go-api/.env.example`):

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `PORT` | No | `8080` | Fiber listen port |
| `NODE_API_URL` | Yes | — | Absolute HTTP(S) base URL for the Node API |
| `CORS_ALLOWED_ORIGINS` | No | empty value | Comma-separated allowed origins |

The outbound statistics client has a fixed five-second HTTP timeout. There is
currently no timeout environment variable; changing that is a separate
configuration decision.

## Reusable design notes

`usecase.Service` owns the only application operation: QR factorization runs
first, and statistics are requested only after QR succeeds. Its statistics
collaborator is a small interface, which keeps orchestration tests independent
of an HTTP server. `handler` owns HTTP concerns exclusively, while `qr` remains
pure and independently testable. `statistics.Result` and `qr.Matrix` are
calculation/transfer values, not persistent domain entities.

From `go-api/`, format and validate with:

```bash
find . -type f -name '*.go' -exec gofmt -w {} +
go vet ./...
go test ./...
go build ./...
```
