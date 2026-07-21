package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sensor-telemetry-service/backend/internal/telemetry"
)

const insertReadingQuery = `
INSERT INTO readings (sensor_id, recorded_at, value, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (sensor_id, recorded_at) DO NOTHING`

const findReadingQuery = `
SELECT recorded_at, value, status::text
FROM readings
WHERE sensor_id = $1 AND recorded_at = $2`

const listReadingsQuery = `
SELECT recorded_at, value, status::text
FROM readings
WHERE sensor_id = $1
  AND recorded_at >= $2
  AND recorded_at < $3
ORDER BY recorded_at ASC`

func (s *Store) ListReadings(
	ctx context.Context,
	sensorID string,
	from time.Time,
	to time.Time,
) ([]telemetry.Reading, error) {
	rows, err := s.pool.Query(ctx, listReadingsQuery, sensorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()

	readings := make([]telemetry.Reading, 0)
	for rows.Next() {
		var reading telemetry.Reading
		var status string
		if err := rows.Scan(&reading.RecordedAt, &reading.Value, &status); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		reading.Status = telemetry.ReadingStatus(status)
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate readings: %w", err)
	}

	return readings, nil
}

func (s *Store) InsertReadings(
	ctx context.Context,
	sensorID string,
	readings []telemetry.Reading,
) ([]telemetry.WriteResult, error) {
	if len(readings) == 0 {
		return []telemetry.WriteResult{}, nil
	}

	// Save all valid items together. If the database fails, save none of them.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin reading transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	results := make([]telemetry.WriteResult, 0, len(readings))
	for _, reading := range readings {
		tag, err := tx.Exec(
			ctx,
			insertReadingQuery,
			sensorID,
			reading.RecordedAt,
			reading.Value,
			reading.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("insert reading: %w", err)
		}

		if tag.RowsAffected() == 1 {
			results = append(results, telemetry.WriteResult{Outcome: telemetry.WriteOutcomeStored})
			continue
		}

		// Keep the first reading. Compare values to tell a retry from a conflict.
		existing, err := findReading(ctx, tx, sensorID, reading.RecordedAt)
		if err != nil {
			return nil, err
		}
		outcome := telemetry.WriteOutcomeConflict
		if existing.Value == reading.Value {
			outcome = telemetry.WriteOutcomeDuplicate
		}
		results = append(results, telemetry.WriteResult{
			Outcome:        outcome,
			ExistingValue:  existing.Value,
			ExistingStatus: existing.Status,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit reading transaction: %w", err)
	}
	return results, nil
}

func findReading(
	ctx context.Context,
	tx pgx.Tx,
	sensorID string,
	recordedAt time.Time,
) (telemetry.Reading, error) {
	var reading telemetry.Reading
	var status string
	err := tx.QueryRow(ctx, findReadingQuery, sensorID, recordedAt).Scan(
		&reading.RecordedAt,
		&reading.Value,
		&status,
	)
	if err != nil {
		return telemetry.Reading{}, fmt.Errorf("find existing reading: %w", err)
	}
	reading.Status = telemetry.ReadingStatus(status)
	return reading, nil
}
