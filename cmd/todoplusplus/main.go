package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Harschmann/Todo-/calendar"
	"github.com/Harschmann/Todo-/core" // Add this import back
	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	setupLogging()

	if err := db.Init("tracker.db"); err != nil {
		log.Fatal(err)
	}

	// Add this line back to start the backup service
	go core.StartPeriodicBackups()

	calendar.Authenticate()

	initialModel := tui.NewForm()

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func setupLogging() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}
