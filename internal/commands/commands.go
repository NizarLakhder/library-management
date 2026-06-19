// Package commands holds the write operations (create / borrow / return) that
// turn the application from a read-only dashboard into a real management tool.
// Like queries, it imports models but not Fyne, so it stays unit-testable.
// Multi-table writes run inside a transaction so a partial failure rolls back.
package commands

import (
	"errors"
	"fmt"

	"github.com/NizarLakhder/library-management/internal/models"

	"gorm.io/gorm"
)

const (
	statusActif      = "actif"
	statusDisponible = "disponible"
	statusEmprunte   = "emprunte"
)

var errNotConnected = errors.New("base de données non connectée")

// AddAdherent creates a new library member. Status defaults to "actif".
func AddAdherent(db *gorm.DB, nom, prenom, status string) (*models.Adherant, error) {
	if nom == "" || prenom == "" {
		return nil, errors.New("le nom et le prénom de l'abonné sont obligatoires")
	}
	if status == "" {
		status = statusActif
	}
	if db == nil {
		return nil, errNotConnected
	}

	adherent := &models.Adherant{Nom: nom, Prenom: prenom, Status: status}
	if err := db.Create(adherent).Error; err != nil {
		return nil, fmt.Errorf("ajout de l'abonné: %w", err)
	}
	return adherent, nil
}

// AddLivre creates a book, links it to an author (reusing an existing author
// with the same name) and creates nbExemplaires available copies, all in one
// transaction. nbExemplaires below 1 is treated as 1.
func AddLivre(db *gorm.DB, isbn, titre, genre, auteurNom, auteurPrenom string, nbExemplaires int) (*models.LivreInfo, error) {
	if isbn == "" || titre == "" {
		return nil, errors.New("l'ISBN et le titre du livre sont obligatoires")
	}
	if auteurNom == "" {
		return nil, errors.New("le nom de l'auteur est obligatoire")
	}
	if nbExemplaires < 1 {
		nbExemplaires = 1
	}
	if db == nil {
		return nil, errNotConnected
	}

	livre := &models.LivreInfo{Isbn: isbn, Titre: titre, Genre: genre}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(livre).Error; err != nil {
			return fmt.Errorf("création du livre: %w", err)
		}

		// Reuse an author with the same name/first name, otherwise create one.
		var auteur models.Auteur
		err := tx.Where("nom = ? AND prenom = ?", auteurNom, auteurPrenom).First(&auteur).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			auteur = models.Auteur{Nom: auteurNom, Prenom: auteurPrenom}
			if err := tx.Create(&auteur).Error; err != nil {
				return fmt.Errorf("création de l'auteur: %w", err)
			}
		case err != nil:
			return fmt.Errorf("recherche de l'auteur: %w", err)
		}

		if err := tx.Model(livre).Association("Auteurs").Append(&auteur); err != nil {
			return fmt.Errorf("association livre-auteur: %w", err)
		}

		for range nbExemplaires {
			ex := models.Exemplaire{Isbn: isbn, Status: statusDisponible}
			if err := tx.Create(&ex).Error; err != nil {
				return fmt.Errorf("création d'un exemplaire: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livre, nil
}

// BorrowExemplaire records a loan of a copy by a member (date_pret = today),
// refusing copies that already have an open loan, and flips the copy status to
// "emprunte".
func BorrowExemplaire(db *gorm.DB, codeAdherant, exemplaireID int) error {
	if codeAdherant <= 0 || exemplaireID <= 0 {
		return errors.New("le code de l'abonné et l'ID de l'exemplaire doivent être positifs")
	}
	if db == nil {
		return errNotConnected
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var open int64
		if err := tx.Model(&models.Emprunts{}).
			Where("exemplaire_id = ? AND date_retour IS NULL", exemplaireID).
			Count(&open).Error; err != nil {
			return fmt.Errorf("vérification des emprunts en cours: %w", err)
		}
		if open > 0 {
			return fmt.Errorf("l'exemplaire %d est déjà emprunté", exemplaireID)
		}

		// CURRENT_DATE keeps the loan date in the database's own timezone,
		// avoiding any Go/Postgres date mismatch.
		if err := tx.Exec(
			"INSERT INTO emprunts (code_adherant, exemplaire_id, date_pret) VALUES (?, ?, CURRENT_DATE)",
			codeAdherant, exemplaireID).Error; err != nil {
			return fmt.Errorf("création de l'emprunt: %w", err)
		}
		if err := tx.Model(&models.Exemplaire{}).
			Where("exemplaire_id = ?", exemplaireID).
			Update("status", statusEmprunte).Error; err != nil {
			return fmt.Errorf("mise à jour du statut de l'exemplaire: %w", err)
		}
		return nil
	})
}

// ReturnExemplaire closes the open loan of a copy (date_retour = today) and
// flips the copy status back to "disponible". The DATE CHECK forbids returning
// a copy the same day it was borrowed.
func ReturnExemplaire(db *gorm.DB, exemplaireID int) error {
	if exemplaireID <= 0 {
		return errors.New("l'ID de l'exemplaire doit être positif")
	}
	if db == nil {
		return errNotConnected
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Close the open loan only if the return date is strictly after the loan
		// date (the schema's CHECK forbids same-day returns). All date logic stays
		// in SQL via CURRENT_DATE, so there is no Go/Postgres timezone drift.
		res := tx.Exec(
			`UPDATE emprunts SET date_retour = CURRENT_DATE
			 WHERE exemplaire_id = ? AND date_retour IS NULL AND CURRENT_DATE > date_pret`,
			exemplaireID)
		if res.Error != nil {
			return fmt.Errorf("mise à jour de la date de retour: %w", res.Error)
		}

		if res.RowsAffected == 0 {
			// Nothing updated: either no open loan, or it was borrowed today.
			var open int64
			if err := tx.Model(&models.Emprunts{}).
				Where("exemplaire_id = ? AND date_retour IS NULL", exemplaireID).
				Count(&open).Error; err != nil {
				return fmt.Errorf("recherche de l'emprunt: %w", err)
			}
			if open == 0 {
				return fmt.Errorf("aucun emprunt en cours pour l'exemplaire %d", exemplaireID)
			}
			return errors.New("retour impossible le jour même de l'emprunt (la date de retour doit être postérieure)")
		}

		if err := tx.Model(&models.Exemplaire{}).
			Where("exemplaire_id = ?", exemplaireID).
			Update("status", statusDisponible).Error; err != nil {
			return fmt.Errorf("mise à jour du statut de l'exemplaire: %w", err)
		}
		return nil
	})
}
