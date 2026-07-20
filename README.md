# Sensor Telemetry Service

Go and PostgreSQL take-home project for collecting and viewing industrial sensor readings.

## Repository structure

```text
backend/                 Go HTTP service
docker-compose.yml       Local development environment
```

## Start the project

```sh
docker compose up --build
```

- Backend health check: <http://localhost:8080/health>
- PostgreSQL: `localhost:5432`

This first commit only establishes the project structure. Sensor ingestion, queries, summaries, overview behavior, and tests will be implemented in separate commits.

## AI usage

OpenAI Codex was used to read the supplied requirements and create the initial repository structure. All code will be reviewed and understood before submission.
