package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Harschmann/Todo-/calendar"
	"github.com/Harschmann/Todo-/core"
	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	appDataDir, err := db.GetAppDataDir()
	if err != nil {
		fmt.Printf("Fatal error: could not get app data directory: %v", err)
		os.Exit(1)
	}

	reminderFlag := flag.Bool("reminder", false, "Send a reminder email if no log is present for today.")
	exportFlag := flag.Bool("export", false, "Export all logs to an Excel file.")
	flag.Parse()

	setupLogging(filepath.Join(appDataDir, "app.log"))
	if err := db.Init(filepath.Join(appDataDir, "tracker.db")); err != nil {
		log.Fatal(err)
	}

	if *reminderFlag {
		calendar.Authenticate(appDataDir)
		log.Println("Running in reminder-only mode...")
		core.CheckAndSendReminder()
		log.Println("Reminder check complete.")

	} else if *exportFlag {
		fmt.Println("Exporting logs to Excel...")
		// UPDATED: Pass the data path to the export function
		fileName, err := db.ExportToExcel(appDataDir)
		if err != nil {
			log.Fatalf("Failed to export to Excel: %v", err)
		}
		fmt.Printf("Successfully exported logs to %s\n", fileName)

	} else {
		calendar.Authenticate(appDataDir)
		// UPDATED: Pass the data path to the backup service
		go core.StartPeriodicBackups(appDataDir)

		initialModel := tui.NewForm()
		p := tea.NewProgram(initialModel)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

func setupLogging(logPath string) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}
