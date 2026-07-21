package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

type ingestResponse struct {
	Stored     int            `json:"stored"`
	Duplicates int            `json:"duplicates"`
	Conflicts  int            `json:"conflicts"`
	Rejected   int            `json:"rejected"`
	Results    []ingestResult `json:"results"`
}

type ingestResult struct {
	Index         int                     `json:"index"`
	Outcome       string                  `json:"outcome"`
	Status        telemetry.ReadingStatus `json:"status,omitempty"`
	ExistingValue *float64                `json:"existing_value,omitempty"`
	Error         string                  `json:"error,omitempty"`
}

func (s *Server) ingestReadings(w http.ResponseWriter, r *http.Request) {
	items, err := decodeBatch(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sensor, found, err := s.store.FindSensor(r.Context(), r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "sensor not found"})
		return
	}

	response := ingestResponse{Results: make([]ingestResult, len(items))}
	validReadings := make([]telemetry.Reading, 0, len(items))
	validIndexes := make([]int, 0, len(items))
	for index, item := range items {
		reading, err := decodeReading(item, sensor)
		if err != nil {
			response.Rejected++
			response.Results[index] = ingestResult{Index: index, Outcome: "rejected", Error: err.Error()}
			continue
		}
		validReadings = append(validReadings, reading)
		validIndexes = append(validIndexes, index)
	}
	if len(validReadings) == 0 {
		writeJSON(w, http.StatusOK, response)
		return
	}

	writeResults, err := s.store.InsertReadings(r.Context(), sensor.ID, validReadings)
	if err != nil || len(writeResults) != len(validReadings) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	for resultIndex, writeResult := range writeResults {
		inputIndex := validIndexes[resultIndex]
		result := ingestResult{Index: inputIndex, Outcome: string(writeResult.Outcome)}

		switch writeResult.Outcome {
		case telemetry.WriteOutcomeStored:
			response.Stored++
			result.Status = validReadings[resultIndex].Status
		case telemetry.WriteOutcomeDuplicate:
			response.Duplicates++
			result.Status = writeResult.ExistingStatus
		case telemetry.WriteOutcomeConflict:
			response.Conflicts++
			result.Status = writeResult.ExistingStatus
			existingValue := writeResult.ExistingValue
			result.ExistingValue = &existingValue
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		response.Results[inputIndex] = result
	}

	writeJSON(w, http.StatusOK, response)
}

func decodeBatch(r *http.Request) ([]json.RawMessage, error) {
	decoder := json.NewDecoder(r.Body)
	var items []json.RawMessage
	if err := decoder.Decode(&items); err != nil || items == nil {
		return nil, errors.New("request body must be a JSON array")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, errors.New("request body must contain a single JSON array")
	}
	return items, nil
}

func decodeReading(raw json.RawMessage, sensor telemetry.Sensor) (telemetry.Reading, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return telemetry.Reading{}, errors.New("reading must be a JSON object")
	}

	recordedAtRaw, ok := fields["recorded_at"]
	if !ok || string(recordedAtRaw) == "null" {
		return telemetry.Reading{}, errors.New("recorded_at is required")
	}
	var recordedAtText string
	if err := json.Unmarshal(recordedAtRaw, &recordedAtText); err != nil {
		return telemetry.Reading{}, errors.New("recorded_at must be a string")
	}
	recordedAt, err := time.Parse(time.RFC3339, recordedAtText)
	if err != nil {
		return telemetry.Reading{}, errors.New("recorded_at is not a valid RFC 3339 timestamp")
	}

	valueRaw, ok := fields["value"]
	if !ok || string(valueRaw) == "null" {
		return telemetry.Reading{}, errors.New("value is required")
	}
	var value float64
	if err := json.Unmarshal(valueRaw, &value); err != nil {
		return telemetry.Reading{}, errors.New("value must be a number")
	}

	return telemetry.Reading{
		RecordedAt: recordedAt.UTC(),
		Value:      value,
		Status:     telemetry.ClassifyReading(sensor, value),
	}, nil
}
