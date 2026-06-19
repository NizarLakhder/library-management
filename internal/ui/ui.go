// Package ui builds and runs the Fyne desktop window. It owns the connection
// form, the query buttons and the results table, delegating the actual database
// work to the database and queries packages.
package ui

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"gorm.io/gorm"

	"github.com/NizarLakhder/library-management/internal/commands"
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

	// runQuery executes a report and renders it. currentQuery remembers the last
	// one so the table can be refreshed after a write operation.
	var currentQuery *queries.Query
	runQuery := func(q queries.Query) {
		newData, err := q.Execute(db)
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
			maxColsToSet := min(numColsData, len(q.ColumnWidths))
			for i := range maxColsToSet {
				resultsTable.SetColumnWidth(i, q.ColumnWidths[i])
			}
		}
		resultsTable.Refresh()
	}

	for _, q := range qs {
		query := q
		btn := widget.NewButton(query.Label, nil)
		btn.OnTapped = func() {
			highlightActive(btn)
			currentQuery = &query
			runQuery(query)
		}
		btn.Disable()
		queryButtons = append(queryButtons, btn)
		queryButtonBoxItems = append(queryButtonBoxItems, btn)
	}
	queryButtonBox := container.NewHBox(queryButtonBoxItems...)

	// refreshCurrentReport re-runs the last selected report so the table reflects
	// a write that just happened.
	refreshCurrentReport := func() {
		if currentQuery != nil {
			runQuery(*currentQuery)
		}
	}

	// Management (write) actions. The connection is read at click time via the
	// getter so the actions always use the live *gorm.DB.
	actionButtons := newActionButtons(myWindow, func() *gorm.DB { return db }, refreshCurrentReport)
	actionButtonBoxItems := make([]fyne.CanvasObject, len(actionButtons))
	for i, b := range actionButtons {
		actionButtonBoxItems[i] = b
	}
	actionButtonBox := container.NewHBox(actionButtonBoxItems...)

	setButtons := func(enable bool) {
		all := append(append([]*widget.Button{}, queryButtons...), actionButtons...)
		for _, btn := range all {
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
		widget.NewLabel("Consulter (lecture) :"),
		container.NewHScroll(queryButtonBox),
		widget.NewLabel("Gestion (écriture) :"),
		container.NewHScroll(actionButtonBox),
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

// newActionButtons builds the disabled-by-default management buttons (add a
// member, add a book, borrow, return). Each opens a form dialog, runs the
// matching command on the live connection (dbOf) and, on success, calls onChange
// to refresh the displayed report.
func newActionButtons(win fyne.Window, dbOf func() *gorm.DB, onChange func()) []*widget.Button {
	addAdherent := widget.NewButton("Ajouter un abonné", func() {
		nom := widget.NewEntry()
		prenom := widget.NewEntry()
		statut := newEntry("actif", "")
		items := []*widget.FormItem{
			{Text: "Nom", Widget: nom},
			{Text: "Prénom", Widget: prenom},
			{Text: "Statut", Widget: statut},
		}
		dialog.ShowForm("Ajouter un abonné", "Ajouter", "Annuler", items, func(ok bool) {
			if !ok {
				return
			}
			a, err := commands.AddAdherent(dbOf(),
				strings.TrimSpace(nom.Text), strings.TrimSpace(prenom.Text), strings.TrimSpace(statut.Text))
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation("Abonné ajouté", fmt.Sprintf("Code abonné : %d", a.CodeAdherant), win)
			onChange()
		}, win)
	})

	addLivre := widget.NewButton("Ajouter un livre", func() {
		isbn := widget.NewEntry()
		titre := widget.NewEntry()
		genre := widget.NewEntry()
		auteurNom := widget.NewEntry()
		auteurPrenom := widget.NewEntry()
		nbExemplaires := newEntry("1", "1")
		items := []*widget.FormItem{
			{Text: "ISBN", Widget: isbn},
			{Text: "Titre", Widget: titre},
			{Text: "Genre", Widget: genre},
			{Text: "Auteur (nom)", Widget: auteurNom},
			{Text: "Auteur (prénom)", Widget: auteurPrenom},
			{Text: "Nb exemplaires", Widget: nbExemplaires},
		}
		dialog.ShowForm("Ajouter un livre", "Ajouter", "Annuler", items, func(ok bool) {
			if !ok {
				return
			}
			nb, _ := strconv.Atoi(strings.TrimSpace(nbExemplaires.Text))
			_, err := commands.AddLivre(dbOf(),
				strings.TrimSpace(isbn.Text), strings.TrimSpace(titre.Text), strings.TrimSpace(genre.Text),
				strings.TrimSpace(auteurNom.Text), strings.TrimSpace(auteurPrenom.Text), nb)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation("Livre ajouté", "Le livre et ses exemplaires ont été créés.", win)
			onChange()
		}, win)
	})

	borrow := widget.NewButton("Emprunter", func() {
		code := widget.NewEntry()
		exID := widget.NewEntry()
		items := []*widget.FormItem{
			{Text: "Code abonné", Widget: code},
			{Text: "ID exemplaire", Widget: exID},
		}
		dialog.ShowForm("Enregistrer un emprunt", "Emprunter", "Annuler", items, func(ok bool) {
			if !ok {
				return
			}
			c, errCode := strconv.Atoi(strings.TrimSpace(code.Text))
			e, errEx := strconv.Atoi(strings.TrimSpace(exID.Text))
			if errCode != nil || errEx != nil {
				dialog.ShowError(errors.New("le code abonné et l'ID exemplaire doivent être des nombres"), win)
				return
			}
			if err := commands.BorrowExemplaire(dbOf(), c, e); err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation("Emprunt enregistré", "L'emprunt a été créé.", win)
			onChange()
		}, win)
	})

	returnCopy := widget.NewButton("Retourner", func() {
		exID := widget.NewEntry()
		items := []*widget.FormItem{
			{Text: "ID exemplaire", Widget: exID},
		}
		dialog.ShowForm("Retourner un exemplaire", "Retourner", "Annuler", items, func(ok bool) {
			if !ok {
				return
			}
			e, err := strconv.Atoi(strings.TrimSpace(exID.Text))
			if err != nil {
				dialog.ShowError(errors.New("l'ID exemplaire doit être un nombre"), win)
				return
			}
			if err := commands.ReturnExemplaire(dbOf(), e); err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation("Retour enregistré", "L'exemplaire a été retourné.", win)
			onChange()
		}, win)
	})

	buttons := []*widget.Button{addAdherent, addLivre, borrow, returnCopy}
	for _, b := range buttons {
		b.Disable()
	}
	return buttons
}
