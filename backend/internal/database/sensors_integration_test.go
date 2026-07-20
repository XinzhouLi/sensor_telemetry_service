package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestListSensorsIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(adminPool.Close)

	schemaName := "sensor_test_" + randomHex(t, 8)
	schemaIdentifier := pgx.Identifier{schemaName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schemaIdentifier); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA "+schemaIdentifier+" CASCADE")
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	executeSQLFile(t, ctx, pool, filepath.Join("..", "..", "..", "database", "migrations", "001_create_schema.sql"))
	executeSQLFile(t, ctx, pool, filepath.Join("..", "..", "..", "database", "seeds", "development.sql"))

	sensors, err := NewStore(pool).ListSensors(ctx)
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

func executeSQLFile(t *testing.T, ctx context.Context, pool *pgxpool.Pool, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if _, err := pool.Exec(ctx, string(contents), pgx.QueryExecModeSimpleProtocol); err != nil {
		t.Fatalf("execute %s: %v", path, err)
	}
}

func randomHex(t *testing.T, byteCount int) string {
	t.Helper()
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		t.Fatalf("generate schema name: %v", err)
	}
	return hex.EncodeToString(value)
}
