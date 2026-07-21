package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

func TestListReadings(t *testing.T) {
	store := &fakeStore{
		findSensor:      telemetry.Sensor{ID: "nox-analyzer-1"},
		findSensorFound: true,
		readings: []telemetry.Reading{
			{
				RecordedAt: time.Date(2026, time.July, 20, 8, 3, 0, 0, time.FixedZone("MDT", -6*60*60)),
				Value:      41.2,
				Status:     telemetry.ReadingStatusValid,
			},
			{
				RecordedAt: time.Date(2026, time.July, 20, 8, 4, 0, 0, time.FixedZone("MDT", -6*60*60)),
				Value:      512,
				Status:     telemetry.ReadingStatusOutOfRange,
			},
		},
	}
	path := "/sensors/nox-analyzer-1/readings?from=2026-07-20T08:00:00-06:00&to=2026-07-20T09:00:00-06:00"
	response := performRequest(t, store, time.Now(), http.MethodGet, path)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var body []readingResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 2 {
		t.Fatalf("len(response) = %d, want 2", len(body))
	}
	if body[0].RecordedAt != "2026-07-20T14:03:00Z" || body[1].RecordedAt != "2026-07-20T14:04:00Z" {
		t.Fatalf("recorded_at values = %q, %q", body[0].RecordedAt, body[1].RecordedAt)
	}
	if body[0].Status != telemetry.ReadingStatusValid || body[1].Status != telemetry.ReadingStatusOutOfRange {
		t.Fatalf("statuses = %q, %q", body[0].Status, body[1].Status)
	}
	wantFrom := time.Date(2026, time.July, 20, 14, 0, 0, 0, time.UTC)
	wantTo := wantFrom.Add(time.Hour)
	if store.readSensorID != "nox-analyzer-1" || !store.readFrom.Equal(wantFrom) || !store.readTo.Equal(wantTo) {
		t.Fatalf("query received sensor=%q from=%s to=%s", store.readSensorID, store.readFrom, store.readTo)
	}
}

func TestListReadingsInvalidWindow(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantBody string
	}{
		{name: "missing from", query: "to=2026-07-20T15:00:00Z", wantBody: `{"error":"from is required"}`},
		{name: "missing to", query: "from=2026-07-20T14:00:00Z", wantBody: `{"error":"to is required"}`},
		{name: "invalid from", query: "from=bad&to=2026-07-20T15:00:00Z", wantBody: `{"error":"from must be a valid RFC 3339 timestamp"}`},
		{name: "invalid to", query: "from=2026-07-20T14:00:00Z&to=bad", wantBody: `{"error":"to must be a valid RFC 3339 timestamp"}`},
		{name: "equal", query: "from=2026-07-20T14:00:00Z&to=2026-07-20T14:00:00Z", wantBody: `{"error":"from must be before to"}`},
		{name: "reversed", query: "from=2026-07-20T15:00:00Z&to=2026-07-20T14:00:00Z", wantBody: `{"error":"from must be before to"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{findSensorErr: errors.New("must not be called")}
			path := "/sensors/unknown/readings?" + test.query
			response := performRequest(t, store, time.Now(), http.MethodGet, path)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", response.Code)
			}
			if got := strings.TrimSpace(response.Body.String()); got != test.wantBody {
				t.Fatalf("body = %s, want %s", got, test.wantBody)
			}
		})
	}
}

func TestListReadingsUnknownSensor(t *testing.T) {
	path := "/sensors/unknown/readings?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	response := performRequest(t, &fakeStore{}, time.Now(), http.MethodGet, path)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestListReadingsEmptyArray(t *testing.T) {
	store := &fakeStore{findSensorFound: true}
	path := "/sensors/nox-analyzer-1/readings?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	response := performRequest(t, store, time.Now(), http.MethodGet, path)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := strings.TrimSpace(response.Body.String()); got != "[]" {
		t.Fatalf("body = %s, want []", got)
	}
}

func TestListReadingsDatabaseErrors(t *testing.T) {
	path := "/sensors/nox-analyzer-1/readings?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	tests := []struct {
		name  string
		store *fakeStore
	}{
		{name: "find sensor", store: &fakeStore{findSensorErr: errors.New("database unavailable")}},
		{
			name: "list readings",
			store: &fakeStore{
				findSensorFound: true,
				readErr:         errors.New("database unavailable"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(t, test.store, time.Now(), http.MethodGet, path)
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", response.Code)
			}
			if got := strings.TrimSpace(response.Body.String()); got != `{"error":"internal server error"}` {
				t.Fatalf("body = %s", got)
			}
		})
	}
}
