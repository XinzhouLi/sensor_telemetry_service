package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"sensor-telemetry-service/backend/internal/telemetry"
)

const listSensorsQuery = `
SELECT
    sensor.id,
    sensor.name,
    sensor.unit,
    sensor.valid_min,
    sensor.valid_max,
    latest.recorded_at,
    latest.value,
    latest.status::text
FROM sensors AS sensor
LEFT JOIN LATERAL (
    SELECT recorded_at, value, status
    FROM readings
    WHERE sensor_id = sensor.id
    ORDER BY recorded_at DESC
    LIMIT 1
) AS latest ON true
ORDER BY sensor.id`

func (s *Store) ListSensors(ctx context.Context) ([]telemetry.Sensor, error) {
	rows, err := s.pool.Query(ctx, listSensorsQuery)
	if err != nil {
		return nil, fmt.Errorf("query sensors: %w", err)
	}
	defer rows.Close()

	sensors := make([]telemetry.Sensor, 0)
	for rows.Next() {
		var sensor telemetry.Sensor
		var recordedAt pgtype.Timestamptz
		var value pgtype.Float8
		var status pgtype.Text

		if err := rows.Scan(
			&sensor.ID,
			&sensor.Name,
			&sensor.Unit,
			&sensor.ValidMin,
			&sensor.ValidMax,
			&recordedAt,
			&value,
			&status,
		); err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}

		if recordedAt.Valid != value.Valid || recordedAt.Valid != status.Valid {
			return nil, fmt.Errorf("scan sensor %q: latest reading contains inconsistent null fields", sensor.ID)
		}
		if recordedAt.Valid {
			sensor.LatestReading = &telemetry.Reading{
				RecordedAt: recordedAt.Time,
				Value:      value.Float64,
				Status:     telemetry.ReadingStatus(status.String),
			}
		}

		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sensors: %w", err)
	}

	return sensors, nil
}
