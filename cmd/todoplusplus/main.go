package main

import (
	"flag"
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
	// 1. Define all command-line flags
	reminderFlag := flag.Bool("reminder", false, "Send a reminder email if no log is present for today.")
	exportFlag := flag.Bool("export", false, "Export all logs to an Excel file.") // ADDED
	flag.Parse()

	// 2. Perform setup common to all modes
	setupLogging()
	if err := db.Init("tracker.db"); err != nil {
		log.Fatal(err)
	}

	// 3. Check which mode to run in
	if *reminderFlag {
		// --- Reminder Mode ---
		calendar.Authenticate() // Auth is needed for reminders
		log.Println("Running in reminder-only mode...")
		core.CheckAndSendReminder()
		log.Println("Reminder check complete.")

	} else if *exportFlag {
		// --- Export Mode ---
		fmt.Println("Exporting logs to Excel...")
		fileName, err := db.ExportToExcel()
		if err != nil {
			log.Fatalf("Failed to export to Excel: %v", err)
		}
		fmt.Printf("Successfully exported logs to %s\n", fileName)

	} else {
		// --- TUI Mode ---
		calendar.Authenticate() // Auth is needed for TUI
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
