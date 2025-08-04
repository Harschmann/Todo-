package main

import (
	"flag" // Add this import
	"fmt"
	"log"
	"os"

	"github.com/Harschmann/Todo-/calendar"
	"github.com/Harschmann/Todo-/core"
	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 1. Define and parse the --reminder flag
	reminderFlag := flag.Bool("reminder", false, "Send a reminder email if no log is present for today.")
	flag.Parse()

	// 2. Perform setup common to both modes
	setupLogging()
	if err := db.Init("tracker.db"); err != nil {
		log.Fatal(err)
	}
	calendar.Authenticate()

	// 3. Check which mode to run in
	if *reminderFlag {
		// --- Reminder Mode ---
		log.Println("Running in reminder-only mode...")
		core.CheckAndSendReminder()
		log.Println("Reminder check complete.")
	} else {
		// --- TUI Mode ---
		// Start background services
		go core.StartPeriodicBackups()

		// Start the TUI
		initialModel := tui.NewForm()
		p := tea.NewProgram(initialModel)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

func setupLogging() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}
