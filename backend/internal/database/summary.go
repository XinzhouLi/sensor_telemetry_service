package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"sensor-telemetry-service/backend/internal/telemetry"
)

// Group readings by UTC hour in PostgreSQL so Go gets only the final summary.
const summarizeReadingsQuery = `
SELECT
    date_trunc('hour', recorded_at, 'UTC') AS bucket_start,
    avg(value) FILTER (WHERE status = 'valid') AS average,
    min(value) FILTER (WHERE status = 'valid') AS minimum,
    max(value) FILTER (WHERE status = 'valid') AS maximum,
    count(*) FILTER (WHERE status = 'valid') AS valid_count,
    count(*) FILTER (WHERE status = 'out_of_range') AS out_of_range_count
FROM readings
WHERE sensor_id = $1
  AND recorded_at >= $2
  AND recorded_at < $3
GROUP BY bucket_start
ORDER BY bucket_start ASC`

func (s *Store) SummarizeReadings(
	ctx context.Context,
	sensorID string,
	from time.Time,
	to time.Time,
) ([]telemetry.SummaryBucket, error) {
	rows, err := s.pool.Query(ctx, summarizeReadingsQuery, sensorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("summarize readings: %w", err)
	}
	defer rows.Close()

	buckets := make([]telemetry.SummaryBucket, 0)
	for rows.Next() {
		var bucket telemetry.SummaryBucket
		var average, minimum, maximum pgtype.Float8
		if err := rows.Scan(
			&bucket.BucketStart,
			&average,
			&minimum,
			&maximum,
			&bucket.ValidCount,
			&bucket.OutOfRangeCount,
		); err != nil {
			return nil, fmt.Errorf("scan reading summary: %w", err)
		}
		// These values are nil when the hour has no valid readings.
		bucket.Average = nullableFloat(average)
		bucket.Minimum = nullableFloat(minimum)
		bucket.Maximum = nullableFloat(maximum)
		buckets = append(buckets, bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reading summaries: %w", err)
	}

	return buckets, nil
}

func nullableFloat(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}
