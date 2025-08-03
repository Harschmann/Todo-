package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Harschmann/Todo-/calendar" // Add this
	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := db.Init("tracker.db"); err != nil {
		log.Fatal(err)
	}

	// First, run the one-time authentication flow if necessary.
	calendar.Authenticate()

	// Now that authentication is done, start the TUI.
	initialModel := tui.NewForm()

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
