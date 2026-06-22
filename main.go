// Command bibliotheque is a desktop library management application.
// It connects to a PostgreSQL database to browse analytical reports (overdue
// loans, popular authors, average loan duration, …) and to manage members,
// books and loans, all in a Fyne UI.
package main

import (
	"github.com/NizarLakhder/library-management/internal/queries"
	"github.com/NizarLakhder/library-management/internal/ui"
)

func main() {
	ui.Run(queries.All)
}
