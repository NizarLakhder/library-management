//go:build integration

// Integration tests for the analytical queries. They run only when built with
// the "integration" build tag against a real PostgreSQL instance:
//
//	TEST_DB_HOST=localhost TEST_DB_PORT=5432 TEST_DB_USER=postgres \
//	TEST_DB_PASSWORD=postgres TEST_DB_NAME=bibliotheque \
//	go test -tags integration ./internal/queries/
//
// They are excluded from the default `go test ./...` run so the unit suite stays
// hermetic.
package queries_test

import (
	"os"
	"testing"

	"github.com/NizarLakhder/library-management/internal/database"
	"github.com/NizarLakhder/library-management/internal/queries"
)

func testConfig(t *testing.T) database.Config {
	t.Helper()
	cfg := database.Config{
		Host:     os.Getenv("TEST_DB_HOST"),
		Port:     os.Getenv("TEST_DB_PORT"),
		User:     os.Getenv("TEST_DB_USER"),
		Password: os.Getenv("TEST_DB_PASSWORD"),
		DBName:   os.Getenv("TEST_DB_NAME"),
	}
	if cfg.Host == "" || cfg.DBName == "" {
		t.Skip("set TEST_DB_HOST and TEST_DB_NAME (plus port/user/password) to run integration tests")
	}
	return cfg
}

// Each query should run without error against the seeded database and return at
// least its header row.
func TestAllQueriesAgainstRealDB(t *testing.T) {
	db, err := database.Connect(testConfig(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer database.Close(db)

	for _, q := range queries.All {
		t.Run(q.Label, func(t *testing.T) {
			data, err := q.Execute(db)
			if err != nil {
				t.Fatalf("Execute(%q): %v", q.Label, err)
			}
			if len(data) < 1 {
				t.Fatalf("Execute(%q): expected at least a header row", q.Label)
			}
			if len(data[0]) != len(q.ColumnWidths) {
				t.Errorf("Execute(%q): header has %d columns but %d column widths declared",
					q.Label, len(data[0]), len(q.ColumnWidths))
			}
		})
	}
}
