package commands

import (
	"strings"
	"testing"
)

// All commands must validate their input before touching the database, so an
// invalid call with a nil handle returns the validation error (not the
// connection error), and a valid call with a nil handle returns the connection
// error.

func TestAddAdherentValidation(t *testing.T) {
	if _, err := AddAdherent(nil, "", "John", "actif"); err == nil {
		t.Error("empty nom should fail validation")
	}
	if _, err := AddAdherent(nil, "Doe", "", "actif"); err == nil {
		t.Error("empty prenom should fail validation")
	}
	_, err := AddAdherent(nil, "Doe", "John", "")
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestAddLivreValidation(t *testing.T) {
	if _, err := AddLivre(nil, "", "Titre", "Genre", "Hugo", "Victor", 1); err == nil {
		t.Error("empty isbn should fail validation")
	}
	if _, err := AddLivre(nil, "isbn", "", "Genre", "Hugo", "Victor", 1); err == nil {
		t.Error("empty titre should fail validation")
	}
	if _, err := AddLivre(nil, "isbn", "Titre", "Genre", "", "Victor", 1); err == nil {
		t.Error("empty author name should fail validation")
	}
	_, err := AddLivre(nil, "isbn", "Titre", "Genre", "Hugo", "Victor", 1)
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestBorrowExemplaireValidation(t *testing.T) {
	if err := BorrowExemplaire(nil, 0, 1); err == nil {
		t.Error("non-positive code should fail validation")
	}
	if err := BorrowExemplaire(nil, 1, 0); err == nil {
		t.Error("non-positive exemplaire id should fail validation")
	}
	err := BorrowExemplaire(nil, 1, 1)
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestReturnExemplaireValidation(t *testing.T) {
	if err := ReturnExemplaire(nil, 0); err == nil {
		t.Error("non-positive exemplaire id should fail validation")
	}
	err := ReturnExemplaire(nil, 5)
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestUpdateAdherentValidation(t *testing.T) {
	if err := UpdateAdherent(nil, 0, "Doe", "John", "actif"); err == nil {
		t.Error("non-positive code should fail validation")
	}
	if err := UpdateAdherent(nil, 1, "", "John", "actif"); err == nil {
		t.Error("empty nom should fail validation")
	}
	err := UpdateAdherent(nil, 1, "Doe", "John", "")
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestUpdateLivreValidation(t *testing.T) {
	if err := UpdateLivre(nil, "", "Titre", "Genre"); err == nil {
		t.Error("empty isbn should fail validation")
	}
	if err := UpdateLivre(nil, "isbn", "", "Genre"); err == nil {
		t.Error("empty titre should fail validation")
	}
	err := UpdateLivre(nil, "isbn", "Titre", "Genre")
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestDeleteAdherentValidation(t *testing.T) {
	if err := DeleteAdherent(nil, 0); err == nil {
		t.Error("non-positive code should fail validation")
	}
	err := DeleteAdherent(nil, 1)
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}

func TestDeleteLivreValidation(t *testing.T) {
	if err := DeleteLivre(nil, ""); err == nil {
		t.Error("empty isbn should fail validation")
	}
	err := DeleteLivre(nil, "isbn")
	if err == nil || !strings.Contains(err.Error(), "non connectée") {
		t.Errorf("valid input with nil db should return connection error, got %v", err)
	}
}
