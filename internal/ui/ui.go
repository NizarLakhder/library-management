// Package ui builds and runs the Fyne desktop window. It owns the connection
// form, the query buttons and the results table, delegating the actual database
// work to the database and queries packages.
package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gorm.io/gorm"

	"github.com/NizarLakhder/library-management/internal/database"
	"github.com/NizarLakhder/library-management/internal/queries"
)

// Run builds the window for the given analytical queries and blocks until it is
// closed, then releases the database connection.
func Run(qs []queries.Query) {
	myApp := app.New()
	if r, err := fyne.LoadResourceFromPath("assets/icon.png"); err == nil {
		myApp.SetIcon(r)
	}
	myWindow := myApp.NewWindow("Gestion Bibliothèque - Affichage Tableau (Go/GORM/Fyne)")
	myWindow.Resize(fyne.NewSize(800, 600))

	// Connection handle and table contents are owned by the closures below
	// instead of package-level globals.
	var db *gorm.DB
	tableData := [][]string{{" ", " "}}

	hostEntry := newEntry("localhost", "localhost")
	portEntry := newEntry("5432", "5432")
	userEntry := newEntry("Nom d'utilisateur BD", "")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Mot de passe BD")
	dbNameEntry := newEntry("Nom de la base de données", "")
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
			cell.(*widget.Label).SetText(tableData[id.Row][id.Col])
		},
	)

	queryErrorLabel := widget.NewLabel("")
	queryErrorLabel.Hide()

	// Buttons are disabled until a successful DB connection is established.
	queryButtons := make([]*widget.Button, 0, len(qs))
	queryButtonBoxItems := make([]fyne.CanvasObject, 0, len(qs))

	// highlightActive colours the clicked report button with the theme's primary
	// colour (and resets the others) so the user can see which report is shown.
	highlightActive := func(active *widget.Button) {
		for _, b := range queryButtons {
			if b == active {
				b.Importance = widget.HighImportance
			} else {
				b.Importance = widget.MediumImportance
			}
			b.Refresh()
		}
	}

	for _, q := range qs {
		query := q
		btn := widget.NewButton(query.Label, nil)
		btn.OnTapped = func() {
			highlightActive(btn)

			newData, err := query.Execute(db)
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
		}
		btn.Disable()
		queryButtons = append(queryButtons, btn)
		queryButtonBoxItems = append(queryButtonBoxItems, btn)
	}
	queryButtonBox := container.NewHBox(queryButtonBoxItems...)

	setButtons := func(enable bool) {
		for _, btn := range queryButtons {
			if enable {
				btn.Enable()
			} else {
				btn.Disable()
			}
		}
	}

	connectButton := widget.NewButton("Connecter", func() {
		statusLabel.SetText("Connexion en cours...")
		cfg := database.Config{
			Host:     strings.TrimSpace(hostEntry.Text),
			Port:     strings.TrimSpace(portEntry.Text),
			User:     strings.TrimSpace(userEntry.Text),
			Password: strings.TrimSpace(passwordEntry.Text),
			DBName:   strings.TrimSpace(dbNameEntry.Text),
		}

		newDB, err := database.Connect(cfg)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Erreur Connexion: %v", err))
			log.Println(err)
			setButtons(false)
			db = nil
			return
		}

		db = newDB
		statusLabel.SetText(fmt.Sprintf("Connecté avec succès à '%s'!", cfg.DBName))
		log.Println("DB connection successful (GORM)!")
		setButtons(true)
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

	myWindow.SetContent(container.NewBorder(topControls, nil, nil, nil, resultsTable))
	myWindow.ShowAndRun()

	if err := database.Close(db); err != nil {
		log.Printf("Error closing database connection: %v", err)
	} else if db != nil {
		log.Println("Database connection closed.")
	}
}

// newEntry returns a text entry with a placeholder and an initial value.
func newEntry(placeholder, value string) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder(placeholder)
	if value != "" {
		e.SetText(value)
	}
	return e
}
