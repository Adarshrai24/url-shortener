# URL Shortener

A full-stack URL shortener built with Go (backend) and vanilla JavaScript (frontend), using PostgreSQL for persistence. Deployed on Render with CI/CD via GitHub Actions.

## Features

- **Shorten URLs**: Convert long URLs into short, shareable links
- **Redirect**: Access original URLs via short keys
- **Duplicate prevention**: Returns existing short URL if the same long URL is submitted
- **Collision-resistant keys**: SHA256 + base64url encoding with counter-based collision resolution
- **Rate limiting**: Limits URL creation and redirects per client IP using a token bucket

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22.6 |
| Database | PostgreSQL (via `lib/pq`) |
| Config | `godotenv` for environment variables |
| Frontend | HTML5, CSS3, Vanilla JavaScript |
| CI/CD | GitHub Actions |
| Deployment | Render |

## Project Structure

```
url_shortener/
├── backend/
│   ├── main.go                 # Entry point, routes, static file serving
│   ├── go.mod / go.sum         # Go module definition
│   ├── .env                    # Environment variables (not committed)
│   ├── db/
│   │   └── db.go               # Database connection pool
│   ├── handlers/
│   │   ├── url.go              # HTTP handlers: ShortenURL, GetURL
│   │   └── url_test.go         # Handler tests
│   ├── middleware/
│   │   ├── rate_limiter.go     # Per-client-IP rate-limiting middleware
│   │   └── rate_limiter_test.go
│   ├── models/
│   │   ├── url.go              # Request/response models
│   │   └── url_test.go         # Model tests
│   └── utils/
│       ├── hash.go             # SHA256 + base64url key generation (7 chars)
│       ├── hash_test.go        # Hash utility tests
│       ├── rate_limiter.go     # In-memory token-bucket implementation
│       └── rate_limiter_test.go
├── frontend/
│   ├── index.html              # Main HTML page
│   ├── index.js                # Client-side logic (fetch, copy, delete)
│   └── styles.css              # Styling
└── .github/workflows/
    └── main.yml                # CI/CD pipeline
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/` | Create a short URL. Body: `{"url": "https://example.com"}` |
| `GET` | `/{key}` | Redirect to the original long URL |
| `GET` | `/` | Serve frontend HTML |
| `GET` | `/static/*` | Serve static assets (CSS, JS) |

Rate limiting applies to `POST /` and `GET /{key}`. Each client IP gets a bucket
with a capacity of 60 requests, replenished at 1 request per second. Requests
made without an available token receive `429 Too Many Requests`. The bucket
state is kept in memory per server process, so it resets when the process
restarts and is not shared between multiple server instances.

### Request/Response Examples

**Shorten URL**
```bash
curl -X POST https://your-domain.com/ \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/Adarshrai24"}'
```
```json
{
  "id": 1,
  "key": "aB3xY9z",
  "short_url": "https://your-domain.com/aB3xY9z",
  "long_url": "https://github.com/Adarshrai24"
}
```

**Redirect**
```bash
curl -L https://your-domain.com/aB3xY9z
# Redirects to https://github.com/Adarshrai24
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Node.js 22+ (for frontend syntax checks in CI)

### Environment Variables

Create a `.env` file in `backend/`:

```env
DATABASE_URL=postgres://user:password@localhost:5432/url_shortener?sslmode=disable
BASE_URL=http://localhost:8090
PORT=8090
```

### Run Locally

```bash
# Backend
cd backend
go mod tidy
go run main.go

# Frontend is served automatically at http://localhost:8090
```

### Run Tests

```bash
cd backend
go test ./...
```

### Benchmark

End-to-end HTTP benchmarks were run on 2026-10-26 against the running Go
service with its configured local PostgreSQL database:

| Scenario | Requests/sec | Average | p50 | p95 | Slowest | Responses |
|----------|-------------:|--------:|----:|----:|--------:|-----------|
| `GET /` (frontend) | 2,658.7 | 3.2 ms | 3.7 ms | 6.1 ms | 7.4 ms | 50 x 200 |
| `GET /static/styles.css` | 4,734.3 | 1.9 ms | 0.7 ms | 7.2 ms | 7.5 ms | 50 x 200 |
| `POST /` (new URL and concurrent duplicates) | 770.6 | 10.2 ms | 5.1 ms | 31.4 ms | 33.6 ms | 41 x 200, 8 x 201, 1 x 500 |
| `POST /` (existing URL) | 500.0 | 16.8 ms | 7.7 ms | 73.6 ms | 77.5 ms | 50 x 200 |
| `GET /{key}` (redirect followed to frontend) | 381.6 | 21.4 ms | 10.3 ms | 80.2 ms | 86.8 ms | 50 x 200 |

Each scenario sent 50 requests with 10 concurrent workers using `hey` and
keep-alive. A fresh server process was used for each rate-limited scenario to
start with a full per-IP bucket. The redirect benchmark follows the `302` to
the local frontend, so its timing includes both requests. The URL-creation
workload used one URL across concurrent requests: it created the URL once,
then exercised duplicate handling. One request returned `500` under that
concurrent create/duplicate workload; this is included in the results rather
than excluded.

Environment: Go 1.22.6, Linux/amd64, Intel Core i3-1005G1 @ 1.20 GHz (4 logical
CPUs), and PostgreSQL 14 on the same machine. These are local, short-run
measurements; database/network placement, system load, and the rate limiter
can materially change production performance.

To reproduce, install [`hey`](https://github.com/rakyll/hey), build and run the
service from `backend/` with a test database configured in `.env`:

```bash
cd backend
go build -o url-shortener-bench .
PORT=18091 BASE_URL=http://127.0.0.1:18091 ./url-shortener-bench
```

In another terminal, benchmark the frontend and static asset:

```bash
hey -n 50 -c 10 http://127.0.0.1:18091/
hey -n 50 -c 10 http://127.0.0.1:18091/static/styles.css
```

For the API, create a unique test URL and use the returned `key` for the
redirect request. The first `hey` run below starts with a new URL, so it
exercises creation and concurrent duplicate handling. Restart the service
between API batches to reset its per-process rate limiter:

```bash
BENCH_URL="http://127.0.0.1:18091/?benchmark=$(date +%s)-$RANDOM"
BODY=$(printf '{"url":"%s"}' "$BENCH_URL")
hey -n 50 -c 10 -m POST \
  -H 'Content-Type: application/json' -d "$BODY" \
  http://127.0.0.1:18091/
KEY=$(curl -fsS -X POST http://127.0.0.1:18091/ \
  -H 'Content-Type: application/json' -d "$BODY" |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["key"])')
# Restart the service, then benchmark duplicate creation:
hey -n 50 -c 10 -m POST \
  -H 'Content-Type: application/json' -d "$BODY" \
  http://127.0.0.1:18091/
# Restart the service again, then benchmark the redirect:
hey -n 50 -c 10 "http://127.0.0.1:18091/$KEY"
```

Remove the test URL from the benchmark database after the run (with
`DATABASE_URL` set in the shell):

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
  -c "DELETE FROM public.url WHERE long_url = '$BENCH_URL';"
```

## Deployment

### Render

1. Create a new Web Service on Render
2. Connect your GitHub repository
3. Build command: `cd backend && go build -o url-shortener .`
4. Start command: `cd backend && ./url-shortener`
5. Add environment variables: `DATABASE_URL`, `BASE_URL`, `PORT`
6. Create a PostgreSQL database on Render and link it

### GitHub Actions CI/CD

The workflow (`.github/workflows/main.yml`) runs on every push/PR to `main`:

1. **Backend job**: Format check, `go vet`, tests, build
2. **Frontend job**: JavaScript syntax check
3. **Deploy job**: Triggers Render deploy hook on merge to `main`

Required secret: `RENDER_DEPLOY_HOOK_URL` (from Render service settings)

## Key Implementation Details

### Key Generation (`utils/hash.go`)

- Uses SHA256 hash of `longURL:counter`
- Encodes with base64url (URL-safe, no padding)
- Truncates to 7 characters (~3.5 trillion combinations)
- Increments counter on collision until unique key found

### Duplicate Handling

Before creating a new entry, the handler checks if the `long_url` already exists in the database and returns the existing short URL.

### Rate Limiting

The rate-limiting middleware identifies clients by the IP address parsed from
the request's remote address and applies the token-bucket limit to URL creation
and short-key redirects. A malformed remote address receives `400 Bad Request`.
The limiter's in-memory state is protected for concurrent requests, but is
local to each server process.

### Static File Serving

Go's `http.FileServer` serves the frontend from `../frontend` at `/static/` and the root path serves `index.html`.
