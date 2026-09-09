# Go–Node integration

The Go API keeps ownership of request parsing and QR factorization. After a
successful factorization, it sends the resulting matrices to the existing
Node statistics endpoint and returns the statistics alongside `q` and `r`.

## Configuration

`NODE_API_URL` is required at Go startup. It must be an absolute `http` or
`https` URL without a query string or fragment. The Go service appends
`/api/v1/statistics` to this base URL. The outbound client has a five-second
timeout.

## Service contract

The Go service continues to expose:

```http
POST /qr
Content-Type: application/json
```

The request remains a matrix payload:

```json
{"matrix":[[1,2],[3,4],[5,6]]}
```

On success, the response is `200 OK` and preserves the QR matrices while
adding the Node result:

```json
{
  "q": [[...]],
  "r": [[...]],
  "statistics": {
    "maximum": 4,
    "minimum": 0,
    "sum": 11,
    "average": 1.375,
    "hasDiagonalMatrix": true
  }
}
```

Go calls Node as follows:

```http
POST {NODE_API_URL}/api/v1/statistics
Content-Type: application/json
```

```json
{"q":[[...]],"r":[[...]]}
```

The downstream response must contain all five fields shown above, with finite
numeric aggregate values and no trailing JSON content.

## Error behavior

Existing invalid QR input remains `400` with `{"error":"invalid_request",...}`.

Downstream failures are intentionally separated from client input errors:

| Condition | Status | Response |
| --- | ---: | --- |
| Node unavailable, non-2xx, malformed or invalid JSON | `502` | `{"error":"downstream_error","message":"unable to obtain statistics"}` |
| Outbound timeout or request context deadline | `504` | `{"error":"downstream_timeout","message":"statistics service timed out"}` |
| Unexpected internal failure | `500` | Existing `internal_error` response |

Node response bodies, URLs, matrices, and configuration values are not exposed
in public error responses.
