package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDB *gorm.DB

var tableData [][]string = [][]string{{" ", " "}}

// Adherant represents a library member
type Adherant struct {
	CodeAdherant int        `gorm:"primaryKey;column:code_adherant"`
	Status       string     `gorm:"column:status;not null"`
	Nom          string     `gorm:"column:nom;not null"`
	Prenom       string     `gorm:"column:prenom;not null"`
	Emprunts     []Emprunts `gorm:"foreignKey:CodeAdherant"`
}

func (Adherant) TableName() string { return "adherant" }

// LivreInfo represents a book in the library
type LivreInfo struct {
	Isbn        string       `gorm:"primaryKey;column:isbn"`
	Titre       string       `gorm:"column:titre;not null"`
	Genre       string       `gorm:"column:genre"`
	Auteurs     []*Auteur    `gorm:"many2many:livre_auteur;"`
	Exemplaires []Exemplaire `gorm:"foreignKey:Isbn"`
}

func (LivreInfo) TableName() string { return "livreinfo" }

// Exemplaire represents a physical copy of a book in the library
type Exemplaire struct {
	ExemplaireID int        `gorm:"primaryKey;column:exemplaire_id"`
	Isbn         string     `gorm:"column:isbn;not null"`
	Status       string     `gorm:"column:status;not null"`
	LivreInfo    LivreInfo  `gorm:"foreignKey:Isbn"`
	Emprunts     []Emprunts `gorm:"foreignKey:ExemplaireID"`
}

func (Exemplaire) TableName() string { return "exemplaire" }

// Auteur represents a book author
type Auteur struct {
	AuteurID int          `gorm:"primaryKey;column:auteur_id"`
	Nom      string       `gorm:"column:nom;not null"`
	Prenom   string       `gorm:"column:prenom"`
	Livres   []*LivreInfo `gorm:"many2many:livre_auteur;"`
}

func (Auteur) TableName() string { return "auteur" }

// Emprunts represents a loan record
type Emprunts struct {
	CodeAdherant int        `gorm:"primaryKey;column:code_adherant"`
	ExemplaireID int        `gorm:"primaryKey;column:exemplaire_id"`
	DatePret     time.Time  `gorm:"primaryKey;column:date_pret;type:date"`
	DateRetour   *time.Time `gorm:"column:date_retour;type:date"`
	Adherant     Adherant   `gorm:"foreignKey:CodeAdherant"`
	Exemplaire   Exemplaire `gorm:"foreignKey:ExemplaireID"`
}

func (Emprunts) TableName() string { return "emprunts" }

type AuteurEmpruntsResult struct {
	AuteurID   int
	Nom        string
	Prenom     string
	NbEmprunts int
}

type DureeMoyenneResult struct {
	DureeMoyenne float64
}

type EmpruntParAnResult struct {
	Annee      int
	NbEmprunts int
}

type GenreEmpruntResult struct {
	Genre      string
	NbEmprunts int
}

type LivreEmpruntResult struct {
	Isbn       string
	Titre      string
	NbEmprunts int
}

type SituationAdherantResult struct {
	CodeAdherant   int
	Nom            string
	Prenom         string
	LivresEnCours  int
	LivresEnRetard int
}

// Query bundles a UI label, column widths, and the GORM function to execute.
type Query struct {
	Label string

	ColumnWidths []float32

	Execute func(db *gorm.DB) ([][]string, error)
}

var queries = []Query{
	{
		Label: "Afficher Livres en Retard",

		ColumnWidths: []float32{50, 120, 120, 250, 100},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []struct {
				CodeAdherant int
				Nom          string
				Prenom       string
				Titre        string
				DatePret     time.Time
			}
			fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
			err := db.Model(&Adherant{}).
				Select("adherant.code_adherant, adherant.nom, adherant.prenom, livreinfo.titre, emprunts.date_pret").
				Joins("JOIN emprunts ON emprunts.code_adherant = adherant.code_adherant").
				Joins("JOIN exemplaire ON exemplaire.exemplaire_id = emprunts.exemplaire_id").
				Joins("JOIN livreinfo ON livreinfo.isbn = exemplaire.isbn").
				Where("emprunts.date_retour IS NULL AND emprunts.date_pret < ?", fourteenDaysAgo).
				Distinct().Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (Retard) Error: %v", err)
				return nil, fmt.Errorf("query 'Livres en Retard' failed: %w", err)
			}

			header := []string{"Code", "Nom", "Prénom", "Titre Livre", "Date Prêt"}
			data := [][]string{header}
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
		Label: "Afficher Auteurs Populaires",

		ColumnWidths: []float32{80, 120, 120, 100},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []AuteurEmpruntsResult
			err := db.Model(&Auteur{}).
				Select("auteur.auteur_id, auteur.nom, auteur.prenom, count(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN livre_auteur ON livre_auteur.auteur_id = auteur.auteur_id").
				Joins("JOIN exemplaire ON exemplaire.isbn = livre_auteur.isbn").
				Joins("JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("auteur.auteur_id, auteur.nom, auteur.prenom").
				Order("nb_emprunts DESC").Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (Auteurs) Error: %v", err)
				return nil, fmt.Errorf("query 'Auteurs Populaires' failed: %w", err)
			}

			header := []string{"ID Auteur", "Nom", "Prénom", "Nb Emprunts"}
			data := [][]string{header}
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
		Label: "Afficher Durée Moyenne Emprunt",

		ColumnWidths: []float32{200},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var result DureeMoyenneResult

			err := db.Model(&Emprunts{}).
				Where("date_retour IS NOT NULL").
				Select("AVG(date_retour - date_pret) as duree_moyenne").
				Scan(&result).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "Scan error") || strings.Contains(err.Error(), "converting NULL") {
					log.Printf("GORM Query (Moyenne) - No data or NULL result, returning 0: %v", err)
					result.DureeMoyenne = 0
					err = nil
				} else {
					log.Printf("GORM Query (Moyenne) Error: %v", err)
					return nil, fmt.Errorf("query 'Durée Moyenne' failed: %w", err)
				}
			}

			header := []string{"Durée Moyenne Emprunt (jours)"}
			data := [][]string{header}

			if result.DureeMoyenne > 0 {
				data = append(data, []string{fmt.Sprintf("%.2f", result.DureeMoyenne)})
			} else {
				data = append(data, []string{"N/A (Pas d'emprunts retournés)"})
			}

			return data, nil
		},
	},
	{
		Label: "Afficher Livres Jamais Empruntés",

		ColumnWidths: []float32{150, 350},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []LivreInfo
			err := db.Where("NOT EXISTS (?)",
				db.Select("1").Model(&Emprunts{}).
					Joins("JOIN exemplaire ON exemplaire.exemplaire_id = emprunts.exemplaire_id").
					Where("exemplaire.isbn = livreinfo.isbn"),
			).Find(&results).Error

			if err != nil {
				log.Printf("GORM Query (Jamais) Error: %v", err)
				return nil, fmt.Errorf("query 'Livres Jamais Empruntés' failed: %w", err)
			}

			header := []string{"ISBN", "Titre (Jamais Emprunté)"}
			data := [][]string{header}
			for _, r := range results {
				data = append(data, []string{r.Isbn, r.Titre})
			}
			return data, nil
		},
	},
	{
		Label: "Emprunts par Année",

		ColumnWidths: []float32{100, 150},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []EmpruntParAnResult
			err := db.Model(&Emprunts{}).
				Select("EXTRACT(YEAR FROM date_pret)::int as annee, COUNT(*) as nb_emprunts").
				Group("annee").
				Order("annee DESC").
				Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (ParAn) Error: %v", err)
				return nil, fmt.Errorf("query 'Emprunts par Année' failed: %w", err)
			}

			header := []string{"Année", "Nb Emprunts"}
			data := [][]string{header}
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
		Label: "Répartition par Genre",

		ColumnWidths: []float32{200, 150},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []GenreEmpruntResult
			err := db.Model(&LivreInfo{}).
				Select("COALESCE(livreinfo.genre, 'Non classé') as genre, COUNT(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN exemplaire ON exemplaire.isbn = livreinfo.isbn").
				Joins("JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("livreinfo.genre").
				Order("nb_emprunts DESC").
				Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (Genre) Error: %v", err)
				return nil, fmt.Errorf("query 'Répartition par Genre' failed: %w", err)
			}

			header := []string{"Genre", "Nb Emprunts"}
			data := [][]string{header}
			for _, r := range results {
				data = append(data, []string{r.Genre, fmt.Sprintf("%d", r.NbEmprunts)})
			}
			return data, nil
		},
	},
	{
		Label: "Emprunts par Livre",

		ColumnWidths: []float32{160, 260, 120},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []LivreEmpruntResult
			err := db.Model(&LivreInfo{}).
				Select("livreinfo.isbn, livreinfo.titre, COUNT(emprunts.date_pret) as nb_emprunts").
				Joins("JOIN exemplaire ON exemplaire.isbn = livreinfo.isbn").
				Joins("LEFT JOIN emprunts ON emprunts.exemplaire_id = exemplaire.exemplaire_id").
				Group("livreinfo.isbn, livreinfo.titre").
				Order("nb_emprunts DESC").
				Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (ParLivre) Error: %v", err)
				return nil, fmt.Errorf("query 'Emprunts par Livre' failed: %w", err)
			}

			header := []string{"ISBN", "Titre", "Nb Emprunts"}
			data := [][]string{header}
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
		Label: "Situation des Abonnés",

		ColumnWidths: []float32{60, 120, 120, 120, 120},
		Execute: func(db *gorm.DB) ([][]string, error) {
			if db == nil {
				return nil, errors.New("database not connected")
			}
			var results []SituationAdherantResult
			fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
			err := db.Model(&Adherant{}).
				Select(`adherant.code_adherant, adherant.nom, adherant.prenom,
					COUNT(CASE WHEN emprunts.date_retour IS NULL THEN 1 END) as livres_en_cours,
					COUNT(CASE WHEN emprunts.date_retour IS NULL AND emprunts.date_pret < ? THEN 1 END) as livres_en_retard`,
					fourteenDaysAgo).
				Joins("LEFT JOIN emprunts ON emprunts.code_adherant = adherant.code_adherant").
				Group("adherant.code_adherant, adherant.nom, adherant.prenom").
				Order("adherant.code_adherant").
				Scan(&results).Error

			if err != nil {
				log.Printf("GORM Query (Situation) Error: %v", err)
				return nil, fmt.Errorf("query 'Situation des Abonnés' failed: %w", err)
			}

			header := []string{"Code", "Nom", "Prénom", "En cours", "En retard"}
			data := [][]string{header}
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

func main() {
	myApp := app.New()
	if r, err := fyne.LoadResourceFromPath("assets/icon.png"); err == nil {
		myApp.SetIcon(r)
	}
	myWindow := myApp.NewWindow("Gestion Bibliothèque - Affichage Tableau (Go/GORM/Fyne)")
	myWindow.Resize(fyne.NewSize(800, 600))

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("localhost")
	hostEntry.SetText("localhost")
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("5432")
	portEntry.SetText("5432")
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Nom d'utilisateur BD")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Mot de passe BD")
	dbNameEntry := widget.NewEntry()
	dbNameEntry.SetPlaceHolder("Nom de la base de données")
	statusLabel := widget.NewLabel("Veuillez entrer les informations de connexion et cliquer sur 'Connecter'.")

	resultsTable := widget.NewTable(
		func() (int, int) {
			if len(tableData) == 0 {
				return 0, 0
			}
			return len(tableData), len(tableData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(tableData) || id.Col >= len(tableData[id.Row]) {
				cell.(*widget.Label).SetText("")
				return
			}
			text := tableData[id.Row][id.Col]
			cell.(*widget.Label).SetText(text)
		},
	)

	queryErrorLabel := widget.NewLabel("")
	queryErrorLabel.Hide()

	queryButtons := []*widget.Button{}
	queryButtonBoxItems := []fyne.CanvasObject{}

	// Buttons are disabled until a successful DB connection is established.
	for _, q := range queries {

		query := q

		btn := widget.NewButton(query.Label, func() {

			newData, err := query.Execute(gormDB)

			if err != nil {
				queryErrorLabel.SetText(fmt.Sprintf("Erreur: %v", err))
				queryErrorLabel.Show()
				tableData = [][]string{{""}}
			} else {
				queryErrorLabel.SetText("")
				queryErrorLabel.Hide()
				tableData = newData

				numColsData := 0
				if len(tableData) > 0 {
					numColsData = len(tableData[0])
				}

				maxColsToSet := min(numColsData, len(query.ColumnWidths))

				for i := range maxColsToSet {
					resultsTable.SetColumnWidth(i, query.ColumnWidths[i])
				}

			}

			resultsTable.Refresh()
		})
		btn.Disable()
		queryButtons = append(queryButtons, btn)
		queryButtonBoxItems = append(queryButtonBoxItems, btn)
	}

	queryButtonBox := container.NewHBox(queryButtonBoxItems...)

	connectButton := widget.NewButton("Connecter", func() {
		statusLabel.SetText("Connexion en cours...")
		host := strings.TrimSpace(hostEntry.Text)
		port := strings.TrimSpace(portEntry.Text)
		user := strings.TrimSpace(userEntry.Text)
		password := strings.TrimSpace(passwordEntry.Text)
		dbname := strings.TrimSpace(dbNameEntry.Text)

		if host == "" || port == "" || user == "" || dbname == "" {
			statusLabel.SetText("Erreur : Tous les champs (sauf mot de passe) sont requis.")
			return
		}

		// key=value format handles special characters in passwords (@, :, /, %) safely
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Toronto",
			host, port, user, password, dbname)

		newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		})

		var err error
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
		if err != nil {
			errorMsg := fmt.Sprintf("Erreur Connexion GORM: %v", err)
			statusLabel.SetText(errorMsg)
			log.Println(errorMsg)
			for _, btn := range queryButtons {
				btn.Disable()
			}
			gormDB = nil
			return
		}
		sqlDB, err := gormDB.DB()
		if err != nil {
			errorMsg := fmt.Sprintf("Erreur obtention sql.DB: %v", err)
			statusLabel.SetText(errorMsg)
			log.Println(errorMsg)
			for _, btn := range queryButtons {
				btn.Disable()
			}
			gormDB = nil
			return
		}

		err = sqlDB.Ping()
		if err != nil {
			errorMsg := fmt.Sprintf("Erreur Ping DB: %v", err)
			statusLabel.SetText(errorMsg)
			log.Println(errorMsg)
			for _, btn := range queryButtons {
				btn.Disable()
			}
			gormDB = nil
			return
		}

		statusLabel.SetText(fmt.Sprintf("Connecté avec succès à '%s'!", dbname))
		log.Println("DB connection successful (GORM)!")

		for _, btn := range queryButtons {
			btn.Enable()
		}
	})

	connectionForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Hôte", Widget: hostEntry},
			{Text: "Port", Widget: portEntry},
			{Text: "Utilisateur", Widget: userEntry},
			{Text: "Mot de passe", Widget: passwordEntry},
			{Text: "Base de données", Widget: dbNameEntry},
		},
	}
	connectionForm.SubmitText = ""
	connectionForm.OnSubmit = nil

	topControls := container.NewVBox(
		connectionForm,
		connectButton,
		statusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Choisir une requête :"),
		container.NewHScroll(queryButtonBox),
		queryErrorLabel,
	)

	content := container.NewBorder(topControls, nil, nil, nil, resultsTable)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()

	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err == nil && sqlDB != nil {
			log.Println("Closing database connection...")
			err := sqlDB.Close()
			if err != nil {
				log.Printf("Error closing database connection: %v", err)
			} else {
				log.Println("Database connection closed.")
			}
		}
	}
}
