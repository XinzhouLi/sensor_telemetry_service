package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

func TestDecodeReading(t *testing.T) {
	sensor := telemetry.Sensor{ValidMin: 0, ValidMax: 250}
	tests := []struct {
		name       string
		input      string
		wantError  string
		wantValue  float64
		wantTime   string
		wantStatus telemetry.ReadingStatus
	}{
		{name: "not an object", input: `42`, wantError: "reading must be a JSON object"},
		{name: "null object", input: `null`, wantError: "reading must be a JSON object"},
		{name: "missing timestamp", input: `{"value": 1}`, wantError: "recorded_at is required"},
		{name: "null timestamp", input: `{"recorded_at": null, "value": 1}`, wantError: "recorded_at is required"},
		{name: "numeric timestamp", input: `{"recorded_at": 1, "value": 1}`, wantError: "recorded_at must be a string"},
		{name: "invalid timestamp", input: `{"recorded_at": "not-a-date", "value": 1}`, wantError: "recorded_at is not a valid RFC 3339 timestamp"},
		{name: "missing value", input: `{"recorded_at": "2026-07-20T14:00:00Z"}`, wantError: "value is required"},
		{name: "null value", input: `{"recorded_at": "2026-07-20T14:00:00Z", "value": null}`, wantError: "value is required"},
		{name: "string value", input: `{"recorded_at": "2026-07-20T14:00:00Z", "value": "41.2"}`, wantError: "value must be a number"},
		{
			name:       "valid UTC",
			input:      `{"recorded_at": "2026-07-20T14:00:00Z", "value": 41.2}`,
			wantValue:  41.2,
			wantTime:   "2026-07-20T14:00:00Z",
			wantStatus: telemetry.ReadingStatusValid,
		},
		{
			name:       "offset normalized to UTC",
			input:      `{"recorded_at": "2026-07-20T08:00:00-06:00", "value": 4.12e1}`,
			wantValue:  41.2,
			wantTime:   "2026-07-20T14:00:00Z",
			wantStatus: telemetry.ReadingStatusValid,
		},
		{
			name:       "out of range stored status",
			input:      `{"recorded_at": "2026-07-20T14:00:00Z", "value": 512}`,
			wantValue:  512,
			wantTime:   "2026-07-20T14:00:00Z",
			wantStatus: telemetry.ReadingStatusOutOfRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reading, err := decodeReading(json.RawMessage(test.input), sensor)
			if test.wantError != "" {
				if err == nil || err.Error() != test.wantError {
					t.Fatalf("error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeReading(): %v", err)
			}
			if reading.Value != test.wantValue || reading.Status != test.wantStatus {
				t.Errorf("reading = %#v", reading)
			}
			if got := reading.RecordedAt.Format(time.RFC3339Nano); got != test.wantTime {
				t.Errorf("recorded_at = %q, want %q", got, test.wantTime)
			}
		})
	}
}

func TestIngestReadingsMixedBatch(t *testing.T) {
	sensor := telemetry.Sensor{ID: "nox-analyzer-1", ValidMin: 0, ValidMax: 250}
	store := &fakeStore{
		findSensor:      sensor,
		findSensorFound: true,
		writeResults: []telemetry.WriteResult{
			{Outcome: telemetry.WriteOutcomeStored},
			{Outcome: telemetry.WriteOutcomeStored},
			{Outcome: telemetry.WriteOutcomeDuplicate, ExistingValue: 41.2, ExistingStatus: telemetry.ReadingStatusValid},
			{Outcome: telemetry.WriteOutcomeConflict, ExistingValue: 41.2, ExistingStatus: telemetry.ReadingStatusValid},
		},
	}
	body := `[
        {"recorded_at":"2026-07-20T14:00:00Z","value":41.2},
        {"recorded_at":"2026-07-20T14:01:00Z","value":512},
        {"recorded_at":"not-a-date","value":1},
        {"recorded_at":"2026-07-20T14:02:00Z","value":41.2},
        {"recorded_at":"2026-07-20T14:03:00Z","value":42.8},
        {"recorded_at":"2026-07-20T14:04:00Z","value":"bad"}
    ]`

	response := performJSONRequest(t, store, body)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var result ingestResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Stored != 2 || result.Duplicates != 1 || result.Conflicts != 1 || result.Rejected != 2 {
		t.Fatalf("unexpected counts: %#v", result)
	}
	wantOutcomes := []string{"stored", "stored", "rejected", "duplicate", "conflict", "rejected"}
	for index, want := range wantOutcomes {
		if result.Results[index].Index != index || result.Results[index].Outcome != want {
			t.Errorf("result[%d] = %#v, want outcome %q", index, result.Results[index], want)
		}
	}
	if result.Results[1].Status != telemetry.ReadingStatusOutOfRange {
		t.Errorf("out-of-range status = %q", result.Results[1].Status)
	}
	if result.Results[4].ExistingValue == nil || *result.Results[4].ExistingValue != 41.2 {
		t.Errorf("conflict existing value = %#v", result.Results[4].ExistingValue)
	}
	if store.writtenSensorID != sensor.ID || len(store.writtenReadings) != 4 {
		t.Errorf("store received sensor=%q readings=%d", store.writtenSensorID, len(store.writtenReadings))
	}
}

func TestIngestReadingsRequestErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "empty body", wantStatus: http.StatusBadRequest},
		{name: "top-level null", body: `null`, wantStatus: http.StatusBadRequest},
		{name: "top-level object", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "malformed JSON", body: `[`, wantStatus: http.StatusBadRequest},
		{name: "trailing JSON", body: `[] {}`, wantStatus: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performJSONRequest(t, &fakeStore{}, test.body)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d: %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}

func TestIngestReadingsUnknownSensor(t *testing.T) {
	response := performJSONRequest(t, &fakeStore{}, `[]`)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestIngestReadingsEmptyBatch(t *testing.T) {
	store := &fakeStore{findSensor: telemetry.Sensor{ID: "nox-analyzer-1"}, findSensorFound: true, writeResults: []telemetry.WriteResult{}}
	response := performJSONRequest(t, store, `[]`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if store.writeCalls != 0 || len(store.writtenReadings) != 0 {
		t.Fatalf("empty batch write calls=%d readings=%d", store.writeCalls, len(store.writtenReadings))
	}
	if got := strings.TrimSpace(response.Body.String()); got != `{"stored":0,"duplicates":0,"conflicts":0,"rejected":0,"results":[]}` {
		t.Fatalf("body = %s", got)
	}
}

func TestIngestReadingsAllRejected(t *testing.T) {
	store := &fakeStore{findSensor: telemetry.Sensor{ID: "nox-analyzer-1"}, findSensorFound: true, writeResults: []telemetry.WriteResult{}}
	response := performJSONRequest(t, store, `[{"value":1},{"recorded_at":"bad","value":1}]`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var result ingestResponse
	_ = json.NewDecoder(response.Body).Decode(&result)
	if result.Rejected != 2 || result.Stored != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestIngestReadingsStoreErrors(t *testing.T) {
	t.Run("find sensor error", func(t *testing.T) {
		store := &fakeStore{findSensorErr: errors.New("database unavailable")}
		response := performJSONRequest(t, store, `[]`)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", response.Code)
		}
	})

	t.Run("insert error", func(t *testing.T) {
		store := &fakeStore{
			findSensor:      telemetry.Sensor{ID: "nox-analyzer-1", ValidMin: 0, ValidMax: 250},
			findSensorFound: true,
			writeErr:        errors.New("commit failed"),
		}
		response := performJSONRequest(t, store, `[{"recorded_at":"2026-07-20T14:00:00Z","value":1}]`)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", response.Code)
		}
	})
}

func performJSONRequest(t *testing.T, store Store, body string) *httptest.ResponseRecorder {
	t.Helper()
	handler := newServer(store, time.Now)
	request := httptest.NewRequest(http.MethodPost, "/sensors/nox-analyzer-1/readings", bytes.NewBufferString(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
