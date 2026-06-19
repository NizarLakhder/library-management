// Package queries holds the read-only analytical reports shown in the UI.
// Each Query bundles its UI label, column widths and a self-contained Execute
// function that runs the report and returns a header-plus-rows table. The
// package imports models but not Fyne, so it compiles and tests without CGo.
package queries

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NizarLakhder/library-management/internal/models"

	"gorm.io/gorm"
)

// overdueDays is the loan period (in days) after which an unreturned book is
// considered overdue.
const overdueDays = 14

// Query bundles a UI label, column widths and the GORM function to execute.
// Execute returns the result as a [][]string whose first row is the header.
type Query struct {
	Label        string
	ColumnWidths []float32
	Execute      func(db *gorm.DB) ([][]string, error)
}

// Result DTOs scanned from the aggregate queries below.
type auteurEmpruntsResult struct {
	AuteurID   int
	Nom        string
	Prenom     string
	NbEmprunts int
}

type dureeMoyenneResult struct {
	DureeMoyenne float64
}

type empruntParAnResult struct {
	Annee      int
	NbEmprunts int
}

type genreEmpruntResult struct {
	Genre      string
	NbEmprunts int
}

type livreEmpruntResult struct {
	Isbn       string
	Titre      string
	NbEmprunts int
}

type situationAdherantResult struct {
	CodeAdherant   int
	Nom            string
	Prenom         string
	LivresEnCours  int
	LivresEnRetard int
}

// overdueThreshold returns the cutoff date before which an unreturned loan is
// overdue.
func overdueThreshold() time.Time {
	return time.Now().AddDate(0, 0, -overdueDays)
}

// All is the ordered list of analytical reports exposed by the application.
var All = []Query{
	{
		Label:        "Afficher Livres en Retard",
		ColumnWidths: []float32{50, 120, 120, 250, 100},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []struct {
				CodeAdherant int
				Nom          string
				Prenom       string
				Titre        string
				DatePret     time.Time
			}
			err := db.Model(&models.Adherant{}).
				Select("adherant.code_adherant, adherant.nom, adherant.prenom, livreinfo.titre, emprunts.date_pret").
				Joins("JOIN emprunts ON emprunts.code_adherant = adherant.code_adherant").
				Joins("JOIN exemplaire ON exemplaire.exemplaire_id = emprunts.exemplaire_id").
				Joins("JOIN livreinfo ON livreinfo.isbn = exemplaire.isbn").
				Where("emprunts.date_retour IS NULL AND emprunts.date_pret < ?", overdueThreshold()).
				Distinct().Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (Retard) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Livres en Retard': %w", err)
			}

			data := [][]string{{"Code", "Nom", "Prénom", "Titre Livre", "Date Prêt"}}
			for _, r := range results {
				data = append(data, []string{
					fmt.Sprintf("%d", r.CodeAdherant),
					r.Nom,
					r.Prenom,
					r.Titre,
					r.DatePret.Format("2006-01-02"),
				})
			}
			return data, nil
		},
	},
	{
		Label:        "Afficher Auteurs Populaires",
		ColumnWidths: []float32{80, 120, 120, 100},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []auteurEmpruntsResult
			err := db.Model(&models.Auteur{}).
				Select("auteur.auteur_id, auteur.nom, auteur.prenom, count(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN livre_auteur ON livre_auteur.auteur_id = auteur.auteur_id").
				Joins("JOIN exemplaire ON exemplaire.isbn = livre_auteur.isbn").
				Joins("JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("auteur.auteur_id, auteur.nom, auteur.prenom").
				Order("nb_emprunts DESC").Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (Auteurs) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Auteurs Populaires': %w", err)
			}

			data := [][]string{{"ID Auteur", "Nom", "Prénom", "Nb Emprunts"}}
			for _, r := range results {
				data = append(data, []string{
					fmt.Sprintf("%d", r.AuteurID),
					r.Nom,
					r.Prenom,
					fmt.Sprintf("%d", r.NbEmprunts),
				})
			}
			return data, nil
		},
	},
	{
		Label:        "Afficher Durée Moyenne Emprunt",
		ColumnWidths: []float32{200},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var result dureeMoyenneResult
			err := db.Model(&models.Emprunts{}).
				Where("date_retour IS NOT NULL").
				Select("AVG(date_retour - date_pret) as duree_moyenne").
				Scan(&result).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "Scan error") || strings.Contains(err.Error(), "converting NULL") {
					log.Printf("GORM Query (Moyenne) - No data or NULL result, returning 0: %v", err)
					result.DureeMoyenne = 0
				} else {
					log.Printf("GORM Query (Moyenne) Error: %v", err)
					return nil, fmt.Errorf("échec de la requête 'Durée Moyenne': %w", err)
				}
			}

			data := [][]string{{"Durée Moyenne Emprunt (jours)"}}
			if result.DureeMoyenne > 0 {
				data = append(data, []string{fmt.Sprintf("%.2f", result.DureeMoyenne)})
			} else {
				data = append(data, []string{"N/A (Pas d'emprunts retournés)"})
			}
			return data, nil
		},
	},
	{
		Label:        "Afficher Livres Jamais Empruntés",
		ColumnWidths: []float32{150, 350},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []models.LivreInfo
			err := db.Where("NOT EXISTS (?)",
				db.Select("1").Model(&models.Emprunts{}).
					Joins("JOIN exemplaire ON exemplaire.exemplaire_id = emprunts.exemplaire_id").
					Where("exemplaire.isbn = livreinfo.isbn"),
			).Find(&results).Error
			if err != nil {
				log.Printf("GORM Query (Jamais) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Livres Jamais Empruntés': %w", err)
			}

			data := [][]string{{"ISBN", "Titre (Jamais Emprunté)"}}
			for _, r := range results {
				data = append(data, []string{r.Isbn, r.Titre})
			}
			return data, nil
		},
	},
	{
		Label:        "Emprunts par Année",
		ColumnWidths: []float32{100, 150},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []empruntParAnResult
			err := db.Model(&models.Emprunts{}).
				Select("EXTRACT(YEAR FROM date_pret)::int as annee, COUNT(*) as nb_emprunts").
				Group("annee").
				Order("annee DESC").
				Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (ParAn) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Emprunts par Année': %w", err)
			}

			data := [][]string{{"Année", "Nb Emprunts"}}
			for _, r := range results {
				data = append(data, []string{
					fmt.Sprintf("%d", r.Annee),
					fmt.Sprintf("%d", r.NbEmprunts),
				})
			}
			return data, nil
		},
	},
	{
		Label:        "Répartition par Genre",
		ColumnWidths: []float32{200, 150},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []genreEmpruntResult
			err := db.Model(&models.LivreInfo{}).
				Select("COALESCE(livreinfo.genre, 'Non classé') as genre, COUNT(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN exemplaire ON exemplaire.isbn = livreinfo.isbn").
				Joins("JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("livreinfo.genre").
				Order("nb_emprunts DESC").
				Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (Genre) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Répartition par Genre': %w", err)
			}

			data := [][]string{{"Genre", "Nb Emprunts"}}
			for _, r := range results {
				data = append(data, []string{r.Genre, fmt.Sprintf("%d", r.NbEmprunts)})
			}
			return data, nil
		},
	},
	{
		Label:        "Emprunts par Livre",
		ColumnWidths: []float32{160, 260, 120},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []livreEmpruntResult
			err := db.Model(&models.LivreInfo{}).
				Select("livreinfo.isbn, livreinfo.titre, COUNT(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN exemplaire ON exemplaire.isbn = livreinfo.isbn").
				Joins("LEFT JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("livreinfo.isbn, livreinfo.titre").
				Order("nb_emprunts DESC").
				Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (ParLivre) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Emprunts par Livre': %w", err)
			}

			data := [][]string{{"ISBN", "Titre", "Nb Emprunts"}}
			for _, r := range results {
				data = append(data, []string{
					r.Isbn,
					r.Titre,
					fmt.Sprintf("%d", r.NbEmprunts),
				})
			}
			return data, nil
		},
	},
	{
		Label:        "Situation des Abonnés",
		ColumnWidths: []float32{60, 120, 120, 120, 120},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("base de données non connectée")
			}
			var results []situationAdherantResult
			err := db.Model(&models.Adherant{}).
				Select(`adherant.code_adherant, adherant.nom, adherant.prenom,
					COUNT(CASE WHEN emprunts.date_retour IS NULL THEN 1 END) as livres_en_cours,
					COUNT(CASE WHEN emprunts.date_retour IS NULL AND emprunts.date_pret < ? THEN 1 END) as livres_en_retard`,
					overdueThreshold()).
				Joins("LEFT JOIN emprunts ON emprunts.code_adherant = adherant.code_adherant").
				Group("adherant.code_adherant, adherant.nom, adherant.prenom").
				Order("adherant.code_adherant").
				Scan(&results).Error
			if err != nil {
				log.Printf("GORM Query (Situation) Error: %v", err)
				return nil, fmt.Errorf("échec de la requête 'Situation des Abonnés': %w", err)
			}

			data := [][]string{{"Code", "Nom", "Prénom", "En cours", "En retard"}}
			for _, r := range results {
				data = append(data, []string{
					fmt.Sprintf("%d", r.CodeAdherant),
					r.Nom,
					r.Prenom,
					fmt.Sprintf("%d", r.LivresEnCours),
					fmt.Sprintf("%d", r.LivresEnRetard),
				})
			}
			return data, nil
		},
	},
}
