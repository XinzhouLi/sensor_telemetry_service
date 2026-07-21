# Sensor Telemetry Service

Go and PostgreSQL take-home project for collecting and viewing industrial sensor readings.

## Repository structure

```text
backend/                 Go HTTP service
database/migrations/     PostgreSQL schema changes
database/seeds/          Local demonstration data
docker-compose.yml       Local development environment
```

## Start the project

```sh
docker compose up --build
```

- Backend health check: <http://localhost:8080/health>
- Sensor overview: <http://localhost:8080/sensors>
- Batch reading ingestion: `POST /sensors/{id}/readings`
- PostgreSQL: `localhost:5432`

## Database design

- Schema migrations and local seed data are kept separate.
- Duplicate sensor timestamps keep the first reading; different values are reported as conflicts.
- Reading status is stored at ingestion; sensor health is calculated when queried.
- Docker initializes schema and seed data only for a new database volume.

## AI usage

OpenAI Codex was used to review the requirements and assist with the project structure, database design, API implementation, and tests. All submitted code is reviewed and understood before delivery.
