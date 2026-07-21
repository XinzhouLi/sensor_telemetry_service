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

func (s *Store) InsertReadings(
	ctx context.Context,
	sensorID string,
	readings []telemetry.Reading,
) ([]telemetry.WriteResult, error) {
	if len(readings) == 0 {
		return []telemetry.WriteResult{}, nil
	}

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
