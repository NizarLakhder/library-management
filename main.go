// Command bibliotheque is a desktop library-analytics application.
// It connects to a PostgreSQL database and displays read-only reports
// (overdue loans, popular authors, average loan duration, etc.) in a Fyne UI.
package main

import (
	"github.com/NizarLakhder/library-management/internal/queries"
	"github.com/NizarLakhder/library-management/internal/ui"
)

func main() {
	ui.Run(queries.All)
}
