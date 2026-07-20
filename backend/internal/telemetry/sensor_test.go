package telemetry

import (
	"testing"
	"time"
)

func TestSensorHealth(t *testing.T) {
	now := time.Date(2026, time.July, 20, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		latest *Reading
		want   Health
	}{
		{name: "never reported", latest: nil, want: HealthNeverReported},
		{name: "within healthy window", latest: readingAt(now.Add(-14*time.Minute), ReadingStatusValid), want: HealthOK},
		{name: "exactly at boundary", latest: readingAt(now.Add(-15*time.Minute), ReadingStatusValid), want: HealthOK},
		{name: "older than boundary", latest: readingAt(now.Add(-15*time.Minute-time.Nanosecond), ReadingStatusValid), want: HealthStale},
		{name: "recent out of range", latest: readingAt(now.Add(-time.Minute), ReadingStatusOutOfRange), want: HealthOK},
		{name: "future reading", latest: readingAt(now.Add(time.Minute), ReadingStatusValid), want: HealthOK},
		{
			name:   "equivalent instant in another timezone",
			latest: readingAt(now.Add(-10*time.Minute).In(time.FixedZone("test", -6*60*60)), ReadingStatusValid),
			want:   HealthOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := SensorHealth(test.latest, now); got != test.want {
				t.Fatalf("SensorHealth() = %q, want %q", got, test.want)
			}
		})
	}
}

func readingAt(recordedAt time.Time, status ReadingStatus) *Reading {
	return &Reading{RecordedAt: recordedAt, Status: status}
}
