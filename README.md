# Sensor Telemetry Service

Go and PostgreSQL take-home project for collecting and viewing industrial sensor readings.

## Repository structure

```text
backend/                 Go HTTP service
database/migrations/     PostgreSQL schema changes
database/seeds/          Local demonstration data
frontend/                Flutter Web dashboard served by Nginx
docker-compose.yml       Local development environment
```

- `api` handles HTTP validation and JSON responses.
- `telemetry` contains shared data types and business rules.
- `database` handles PostgreSQL queries and transactions.
- Store and clock dependencies are passed into the API so handlers can be tested without a real server.

Request -> API validation -> telemetry rule or database query -> JSON response

## Run

Docker with Compose support is the only prerequisite.

```sh
docker compose up --build
```

Verify the backend and database connection:

```sh
curl http://localhost:8080/health
```

The response should be `{"status":"ok"}`. PostgreSQL is available on `localhost:5432`, and development seed data makes all three sensor health states visible immediately.

Open `http://localhost:3000` to view the dashboard. It reads sensor overviews through the Nginx `/api` proxy and refreshes them every 15 seconds.

## API

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/sensors` | Every sensor with its latest reading and health |
| `POST` | `/sensors/{id}/readings` | Per-item ingestion outcomes in input order |
| `GET` | `/sensors/{id}/readings?from=...&to=...` | Readings in the window, oldest first |
| `GET` | `/sensors/{id}/summary?from=...&to=...` | UTC hourly summary buckets |

The ingestion body is a JSON array of `recorded_at` and numeric `value` fields. Its response reports `stored`, `duplicate`, `conflict`, or `rejected` for each item. Invalid items do not prevent valid items in the same batch from being processed.

Reading results contain `recorded_at`, `value`, and `status`. Summary buckets contain `bucket_start`, `average`, `minimum`, `maximum`, `valid_count`, and `out_of_range_count`; statistics are `null` when a bucket has no valid readings. A sensor with no readings has `latest_reading: null` and health `never_reported`.

`from` and `to` are required RFC 3339 timestamps. Windows are `[from, to)` and `from >= to` returns `400`. API timestamps are returned in UTC. Unknown sensors return `404`.

## Tests

Run the unit and HTTP handler tests:

```sh
cd backend
go test ./...
```

With the Compose database running, run the PostgreSQL integration tests:

```sh
TEST_DATABASE_URL='postgres://telemetry:telemetry@localhost:5432/telemetry?sslmode=disable' \
  go test ./internal/database -count=1
```

Integration tests create and remove an isolated PostgreSQL schema. Summary bucketing is tested against PostgreSQL because the production behavior uses `date_trunc` and filtered aggregates rather than a duplicate Go implementation.

Run the frontend checks:

```sh
cd frontend
flutter analyze
flutter test
```

## Design decisions

- Schema migrations and local seed data are kept separate.
- Duplicate sensor timestamps keep the first reading. Equal values are idempotent duplicates; different values are reported as conflicts without overwriting history.
- Out-of-range values are stored because abnormal telemetry is operationally important, but only valid readings contribute to summary statistics.
- Reading status is stored at ingestion; sensor health is calculated when queried.
- PostgreSQL performs UTC hourly aggregation so raw windows are not transferred to Go for bucketing.
- Docker initializes schema and seed data only for a new database volume.

## Next steps

- Add pagination, query-window limits, retention policies, and time-based partitioning as data volume grows.
- Introduce a production migration runner instead of relying on first-volume Docker initialization.
- Add structured logs and metrics, plus agreed limits for request size and future sensor timestamps.

## Time spent

Approximately ___ hours.

## AI usage

OpenAI Codex was used to review the requirements and assist with the project structure, database design, API implementation, dashboard, and tests. All submitted code is reviewed and understood before delivery.
