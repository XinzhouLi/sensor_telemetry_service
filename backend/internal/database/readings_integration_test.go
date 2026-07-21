package database

import (
	"context"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

func TestInsertReadingsIntegration(t *testing.T) {
	store, pool := newIntegrationStore(t)
	ctx := context.Background()
	recordedAt := time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC)

	results, err := store.InsertReadings(ctx, "nox-analyzer-1", []telemetry.Reading{
		{RecordedAt: recordedAt, Value: 100, Status: telemetry.ReadingStatusValid},
		{RecordedAt: recordedAt, Value: 100, Status: telemetry.ReadingStatusValid},
		{RecordedAt: recordedAt, Value: 101, Status: telemetry.ReadingStatusValid},
	})
	if err != nil {
		t.Fatalf("InsertReadings(): %v", err)
	}
	wantOutcomes := []telemetry.WriteOutcome{
		telemetry.WriteOutcomeStored,
		telemetry.WriteOutcomeDuplicate,
		telemetry.WriteOutcomeConflict,
	}
	for index, want := range wantOutcomes {
		if results[index].Outcome != want {
			t.Errorf("result[%d] = %q, want %q", index, results[index].Outcome, want)
		}
	}
	if results[2].ExistingValue != 100 || results[2].ExistingStatus != telemetry.ReadingStatusValid {
		t.Errorf("conflict existing reading = %#v", results[2])
	}

	var count int
	var storedValue float64
	if err := pool.QueryRow(
		ctx,
		"SELECT count(*), max(value) FROM readings WHERE sensor_id = $1 AND recorded_at = $2",
		"nox-analyzer-1",
		recordedAt,
	).Scan(&count, &storedValue); err != nil {
		t.Fatalf("query inserted reading: %v", err)
	}
	if count != 1 || storedValue != 100 {
		t.Fatalf("stored count=%d value=%v, want count=1 value=100", count, storedValue)
	}
}

func TestInsertReadingsRollbackIntegration(t *testing.T) {
	store, pool := newIntegrationStore(t)
	ctx := context.Background()
	firstTime := time.Date(2030, time.January, 2, 0, 0, 0, 0, time.UTC)

	_, err := store.InsertReadings(ctx, "nox-analyzer-1", []telemetry.Reading{
		{RecordedAt: firstTime, Value: 100, Status: telemetry.ReadingStatusValid},
		{RecordedAt: firstTime.Add(time.Minute), Value: 101, Status: telemetry.ReadingStatus("invalid_status")},
	})
	if err == nil {
		t.Fatal("InsertReadings() error = nil, want transaction error")
	}

	var count int
	if err := pool.QueryRow(
		ctx,
		"SELECT count(*) FROM readings WHERE sensor_id = $1 AND recorded_at >= $2",
		"nox-analyzer-1",
		firstTime,
	).Scan(&count); err != nil {
		t.Fatalf("query rolled-back readings: %v", err)
	}
	if count != 0 {
		t.Fatalf("readings after rollback = %d, want 0", count)
	}
}
