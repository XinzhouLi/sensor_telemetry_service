package database

import (
	"context"
	"testing"
)

func TestListSensorsIntegration(t *testing.T) {
	store, _ := newIntegrationStore(t)
	sensors, err := store.ListSensors(context.Background())
	if err != nil {
		t.Fatalf("ListSensors(): %v", err)
	}
	if len(sensors) != 3 {
		t.Fatalf("len(sensors) = %d, want 3", len(sensors))
	}
	if sensors[0].ID != "nox-analyzer-1" || sensors[0].LatestReading == nil || sensors[0].LatestReading.Value != 41.2 {
		t.Errorf("unexpected first sensor: %#v", sensors[0])
	}
	if sensors[1].ID != "o2-analyzer-1" || sensors[1].LatestReading == nil {
		t.Errorf("unexpected second sensor: %#v", sensors[1])
	}
	if sensors[2].ID != "stack-temp-1" || sensors[2].LatestReading != nil {
		t.Errorf("unexpected third sensor: %#v", sensors[2])
	}
}

func TestFindSensorIntegration(t *testing.T) {
	store, _ := newIntegrationStore(t)
	ctx := context.Background()

	sensor, found, err := store.FindSensor(ctx, "nox-analyzer-1")
	if err != nil || !found {
		t.Fatalf("FindSensor() found=%v err=%v", found, err)
	}
	if sensor.ValidMin != 0 || sensor.ValidMax != 250 {
		t.Fatalf("sensor range = %v-%v", sensor.ValidMin, sensor.ValidMax)
	}
	if _, found, err := store.FindSensor(ctx, "unknown-sensor"); err != nil || found {
		t.Fatalf("unknown FindSensor() found=%v err=%v", found, err)
	}
}
