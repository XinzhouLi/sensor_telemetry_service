package telemetry

import "time"

type ReadingStatus string

const (
	ReadingStatusValid      ReadingStatus = "valid"
	ReadingStatusOutOfRange ReadingStatus = "out_of_range"
)

type Reading struct {
	RecordedAt time.Time
	Value      float64
	Status     ReadingStatus
}

type WriteOutcome string

const (
	WriteOutcomeStored    WriteOutcome = "stored"
	WriteOutcomeDuplicate WriteOutcome = "duplicate"
	WriteOutcomeConflict  WriteOutcome = "conflict"
)

type WriteResult struct {
	Outcome        WriteOutcome
	ExistingValue  float64
	ExistingStatus ReadingStatus
}

type Sensor struct {
	ID            string
	Name          string
	Unit          string
	ValidMin      float64
	ValidMax      float64
	LatestReading *Reading
}

type Health string

const (
	HealthOK            Health = "ok"
	HealthStale         Health = "stale"
	HealthNeverReported Health = "never_reported"
)

const healthyWindow = 15 * time.Minute

func SensorHealth(latest *Reading, now time.Time) Health {
	if latest == nil {
		return HealthNeverReported
	}
	if !latest.RecordedAt.Before(now.Add(-healthyWindow)) {
		return HealthOK
	}
	return HealthStale
}

func ClassifyReading(sensor Sensor, value float64) ReadingStatus {
	if value < sensor.ValidMin || value > sensor.ValidMax {
		return ReadingStatusOutOfRange
	}
	return ReadingStatusValid
}
