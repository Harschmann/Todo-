package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize the database before starting the UI.
	if err := db.Init("tracker.db"); err != nil {
		log.Fatal(err)
	}

	initialModel := tui.NewForm()

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
