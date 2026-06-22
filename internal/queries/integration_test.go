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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NizarLakhder/library-management/internal/commands"
	"github.com/NizarLakhder/library-management/internal/database"
	"github.com/NizarLakhder/library-management/internal/models"
	"github.com/NizarLakhder/library-management/internal/queries"

	"gorm.io/gorm"
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
			for i, row := range data {
				if len(row) != len(data[0]) {
					t.Errorf("Execute(%q): row %d has %d columns, want %d",
						q.Label, i, len(row), len(data[0]))
				}
			}
		})
	}
}

// runReport executes the report with the given label and returns its rows
// (header included).
func runReport(t *testing.T, db *gorm.DB, label string) [][]string {
	t.Helper()
	for _, q := range queries.All {
		if q.Label == label {
			data, err := q.Execute(db)
			if err != nil {
				t.Fatalf("report %q: %v", label, err)
			}
			return data
		}
	}
	t.Fatalf("report %q not found", label)
	return nil
}

// reportHasISBN reports whether the report lists the given ISBN in its first column.
func reportHasISBN(t *testing.T, db *gorm.DB, label, isbn string) bool {
	t.Helper()
	for _, row := range runReport(t, db, label)[1:] { // skip header
		if len(row) > 0 && row[0] == isbn {
			return true
		}
	}
	return false
}

// TestReportsCorrectnessForAKnownBook asserts the actual contents of two reports
// using an isolated book it creates, so the checks hold regardless of the rest
// of the database.
func TestReportsCorrectnessForAKnownBook(t *testing.T) {
	db, err := database.Connect(testConfig(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer database.Close(db)

	unique := time.Now().UnixNano() % 100000000
	isbn := fmt.Sprintf("QTS-%08d", unique)
	auteurNom := fmt.Sprintf("QAuteur-%08d", unique)
	adherentNom := fmt.Sprintf("QAbonne-%08d", unique)

	defer func() {
		var ids []int
		db.Model(&models.Exemplaire{}).Where("isbn = ?", isbn).Pluck("exemplaire_id", &ids)
		if len(ids) > 0 {
			db.Where("exemplaire_id IN ?", ids).Delete(&models.Emprunts{})
		}
		db.Where("isbn = ?", isbn).Delete(&models.Exemplaire{})
		db.Exec("DELETE FROM livre_auteur WHERE isbn = ?", isbn)
		db.Where("isbn = ?", isbn).Delete(&models.LivreInfo{})
		db.Where("nom = ?", auteurNom).Delete(&models.Auteur{})
		db.Where("nom = ?", adherentNom).Delete(&models.Adherant{})
	}()

	if _, err := commands.AddLivre(db, isbn, "Titre requête", "Test", auteurNom, "P", 1); err != nil {
		t.Fatalf("AddLivre: %v", err)
	}

	// A brand-new book with no loan must be listed as "jamais emprunté".
	if !reportHasISBN(t, db, "Livres Jamais Empruntés", isbn) {
		t.Errorf("new book %s should appear in 'Livres Jamais Empruntés'", isbn)
	}

	// Borrow its copy.
	adherent, err := commands.AddAdherent(db, adherentNom, "Test", "")
	if err != nil {
		t.Fatalf("AddAdherent: %v", err)
	}
	var exemplaire models.Exemplaire
	if err := db.Where("isbn = ?", isbn).First(&exemplaire).Error; err != nil {
		t.Fatalf("retrieve exemplaire: %v", err)
	}
	if err := commands.BorrowExemplaire(db, adherent.CodeAdherant, exemplaire.ExemplaireID); err != nil {
		t.Fatalf("BorrowExemplaire: %v", err)
	}

	// Now it must leave "jamais emprunté" and appear once in "Emprunts par Livre".
	if reportHasISBN(t, db, "Livres Jamais Empruntés", isbn) {
		t.Errorf("borrowed book %s should no longer be in 'Livres Jamais Empruntés'", isbn)
	}
	found := false
	for _, row := range runReport(t, db, "Emprunts par Livre")[1:] {
		if len(row) >= 3 && row[0] == isbn {
			found = true
			if row[2] != "1" {
				t.Errorf("'Emprunts par Livre' count for %s = %q, want \"1\"", isbn, row[2])
			}
		}
	}
	if !found {
		t.Errorf("book %s should appear in 'Emprunts par Livre'", isbn)
	}
}
