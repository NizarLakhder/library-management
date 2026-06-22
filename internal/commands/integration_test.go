//go:build integration

// Integration test for the write operations. Runs only with the "integration"
// build tag against a real PostgreSQL instance and cleans up everything it
// creates:
//
//	TEST_DB_HOST=localhost TEST_DB_PORT=5432 TEST_DB_USER=postgres \
//	TEST_DB_PASSWORD=postgres TEST_DB_NAME=bibliotheque \
//	go test -tags integration ./internal/commands/
package commands_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NizarLakhder/library-management/internal/commands"
	"github.com/NizarLakhder/library-management/internal/database"
	"github.com/NizarLakhder/library-management/internal/models"
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

// TestCRUDCycle exercises the full management flow against the database:
// create a member and a book, borrow the exemplaire, refuse a same-day return,
// return it the next day, update both records, and check the foreign-key-aware
// delete rules (refused while loans exist, allowed for a loan-free member) — and
// cleans everything up.
func TestCRUDCycle(t *testing.T) {
	db, err := database.Connect(testConfig(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer database.Close(db)

	// Keep the ISBN within the realistic varchar(20) bound (a real ISBN-13 is
	// ~17 chars); the 8-digit suffix is enough to avoid clashes between runs.
	unique := time.Now().UnixNano() % 100000000
	isbn := fmt.Sprintf("TST-%08d", unique)
	auteurNom := fmt.Sprintf("AuteurT-%08d", unique)
	adherentNom := fmt.Sprintf("AbonneT-%08d", unique)
	adherentNom2 := fmt.Sprintf("AbonneU-%08d", unique) // loan-free member for the delete happy path

	// Cleanup runs even if the test fails partway through.
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
		db.Where("nom IN ?", []string{adherentNom, adherentNom2}).Delete(&models.Adherant{})
	}()

	// Create a member and a book (with one exemplaire).
	adherent, err := commands.AddAdherent(db, adherentNom, "Test", "")
	if err != nil {
		t.Fatalf("AddAdherent: %v", err)
	}
	if _, err := commands.AddLivre(db, isbn, "Titre de test", "Test", auteurNom, "Prenom", 1); err != nil {
		t.Fatalf("AddLivre: %v", err)
	}

	var exemplaire models.Exemplaire
	if err := db.Where("isbn = ?", isbn).First(&exemplaire).Error; err != nil {
		t.Fatalf("retrieve exemplaire: %v", err)
	}

	// Borrow it.
	if err := commands.BorrowExemplaire(db, adherent.CodeAdherant, exemplaire.ExemplaireID); err != nil {
		t.Fatalf("BorrowExemplaire: %v", err)
	}
	var afterBorrow models.Exemplaire
	db.First(&afterBorrow, exemplaire.ExemplaireID)
	if afterBorrow.Status != "emprunte" {
		t.Errorf("exemplaire status after borrow = %q, want emprunte", afterBorrow.Status)
	}

	// Borrowing again must be refused while the loan is open.
	if err := commands.BorrowExemplaire(db, adherent.CodeAdherant, exemplaire.ExemplaireID); err == nil {
		t.Error("second borrow of an open exemplaire should fail")
	}

	// A same-day return is refused by the DATE CHECK guard.
	if err := commands.ReturnExemplaire(db, exemplaire.ExemplaireID); err == nil || !strings.Contains(err.Error(), "jour même") {
		t.Errorf("same-day return should be refused, got %v", err)
	}

	// Backdate the loan by one day, then the return must succeed.
	yesterday := time.Now().AddDate(0, 0, -1)
	db.Model(&models.Emprunts{}).
		Where("exemplaire_id = ? AND date_retour IS NULL", exemplaire.ExemplaireID).
		Update("date_pret", yesterday)

	if err := commands.ReturnExemplaire(db, exemplaire.ExemplaireID); err != nil {
		t.Fatalf("ReturnExemplaire: %v", err)
	}
	var afterReturn models.Exemplaire
	db.First(&afterReturn, exemplaire.ExemplaireID)
	if afterReturn.Status != "disponible" {
		t.Errorf("exemplaire status after return = %q, want disponible", afterReturn.Status)
	}

	// --- Update ---
	if err := commands.UpdateAdherent(db, adherent.CodeAdherant, adherentNom, "Test", "inactif"); err != nil {
		t.Fatalf("UpdateAdherent: %v", err)
	}
	var updatedMember models.Adherant
	db.First(&updatedMember, adherent.CodeAdherant)
	if updatedMember.Status != "inactif" {
		t.Errorf("member status after update = %q, want inactif", updatedMember.Status)
	}

	if err := commands.UpdateLivre(db, isbn, "Titre modifié", "Test"); err != nil {
		t.Fatalf("UpdateLivre: %v", err)
	}
	var updatedBook models.LivreInfo
	db.First(&updatedBook, "isbn = ?", isbn)
	if updatedBook.Titre != "Titre modifié" {
		t.Errorf("book title after update = %q, want \"Titre modifié\"", updatedBook.Titre)
	}

	// --- Delete refused: both still have loan history ---
	if err := commands.DeleteAdherent(db, adherent.CodeAdherant); err == nil || !strings.Contains(err.Error(), "emprunts") {
		t.Errorf("deleting a member with loans should be refused, got %v", err)
	}
	if err := commands.DeleteLivre(db, isbn); err == nil || !strings.Contains(err.Error(), "emprunts") {
		t.Errorf("deleting a book with borrowed copies should be refused, got %v", err)
	}

	// --- Delete happy path: a loan-free member can be removed ---
	freshMember, err := commands.AddAdherent(db, adherentNom2, "Test", "")
	if err != nil {
		t.Fatalf("AddAdherent (fresh): %v", err)
	}
	if err := commands.DeleteAdherent(db, freshMember.CodeAdherant); err != nil {
		t.Fatalf("DeleteAdherent (fresh): %v", err)
	}
	var remaining int64
	db.Model(&models.Adherant{}).Where("code_adherant = ?", freshMember.CodeAdherant).Count(&remaining)
	if remaining != 0 {
		t.Error("loan-free member should have been deleted")
	}
}
