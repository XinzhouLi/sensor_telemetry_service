package database

import (
	"context"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

func TestSummarizeReadingsIntegration(t *testing.T) {
	store, _ := newIntegrationStore(t)
	ctx := context.Background()
	location := time.FixedZone("MDT", -6*60*60)
	from := time.Date(2032, time.July, 20, 8, 15, 0, 0, location)
	to := time.Date(2032, time.July, 20, 10, 0, 0, 0, location)

	_, err := store.InsertReadings(ctx, "nox-analyzer-1", []telemetry.Reading{
		{RecordedAt: from.Add(-time.Minute), Value: 10, Status: telemetry.ReadingStatusValid},
		{RecordedAt: from, Value: 20, Status: telemetry.ReadingStatusValid},
		{RecordedAt: from.Add(15 * time.Minute), Value: 40, Status: telemetry.ReadingStatusValid},
		{RecordedAt: from.Add(30 * time.Minute), Value: 512, Status: telemetry.ReadingStatusOutOfRange},
		{RecordedAt: time.Date(2032, time.July, 20, 9, 0, 0, 0, location), Value: 700, Status: telemetry.ReadingStatusOutOfRange},
		{RecordedAt: time.Date(2032, time.July, 20, 9, 30, 0, 0, location), Value: 800, Status: telemetry.ReadingStatusOutOfRange},
		{RecordedAt: to, Value: 60, Status: telemetry.ReadingStatusValid},
	})
	if err != nil {
		t.Fatalf("InsertReadings(): %v", err)
	}

	buckets, err := store.SummarizeReadings(ctx, "nox-analyzer-1", from, to)
	if err != nil {
		t.Fatalf("SummarizeReadings(): %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("len(buckets) = %d, want 2", len(buckets))
	}

	first := buckets[0]
	wantFirstStart := time.Date(2032, time.July, 20, 14, 0, 0, 0, time.UTC)
	if !first.BucketStart.Equal(wantFirstStart) {
		t.Errorf("first bucket start = %s, want %s", first.BucketStart, wantFirstStart)
	}
	if first.Average == nil || *first.Average != 30 || first.Minimum == nil || *first.Minimum != 20 || first.Maximum == nil || *first.Maximum != 40 {
		t.Errorf("first bucket statistics = %#v", first)
	}
	if first.ValidCount != 2 || first.OutOfRangeCount != 1 {
		t.Errorf("first bucket counts = valid:%d out_of_range:%d", first.ValidCount, first.OutOfRangeCount)
	}

	second := buckets[1]
	wantSecondStart := wantFirstStart.Add(time.Hour)
	if !second.BucketStart.Equal(wantSecondStart) {
		t.Errorf("second bucket start = %s, want %s", second.BucketStart, wantSecondStart)
	}
	if second.Average != nil || second.Minimum != nil || second.Maximum != nil {
		t.Errorf("out-of-range-only statistics = %#v", second)
	}
	if second.ValidCount != 0 || second.OutOfRangeCount != 2 {
		t.Errorf("second bucket counts = valid:%d out_of_range:%d", second.ValidCount, second.OutOfRangeCount)
	}

	empty, err := store.SummarizeReadings(ctx, "nox-analyzer-1", to.Add(time.Hour), to.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("empty SummarizeReadings(): %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty buckets = %#v, want non-nil empty slice", empty)
	}
}
