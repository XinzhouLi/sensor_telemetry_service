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
- PostgreSQL: `localhost:5432`

## Database design

- Schema migrations and local seed data are kept separate.
- Duplicate sensor timestamps keep the first reading.
- Reading status is stored at ingestion; sensor health is calculated when queried.
- Docker initializes schema and seed data only for a new database volume.

## AI usage

OpenAI Codex was used to read the supplied requirements and create the initial repository structure. All code will be reviewed and understood before submission.
