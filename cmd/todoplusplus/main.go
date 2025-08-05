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
	// ADDED: Run the one-time data migration check at the very start.
	if err := migrateData(); err != nil {
		fmt.Printf("Error during data migration: %v\n", err)
		// We can choose to continue or exit. For now, we'll continue.
	}

	appDataDir, err := db.GetAppDataDir()
	if err != nil {
		fmt.Printf("Fatal error: could not get app data directory: %v\n", err)
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
		fileName, err := db.ExportToExcel()
		if err != nil {
			log.Fatalf("Failed to export to Excel: %v", err)
		}
		fmt.Printf("Successfully exported logs to %s\n", fileName)

	} else {
		calendar.Authenticate(appDataDir)
		go core.StartPeriodicBackups(appDataDir)

		initialModel := tui.NewForm()
		p := tea.NewProgram(initialModel)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

// ADDED: This function checks for old data and moves it to the new location.
func migrateData() error {
	// 1. Get the new and old paths
	newDataPath, err := db.GetAppDataDir()
	if err != nil {
		return err
	}
	oldDataPath, err := os.Getwd() // The old path is just the current directory
	if err != nil {
		return err
	}

	// 2. List of files/folders to move
	filesToMigrate := []string{"tracker.db", "token.json", "app.log", "backups"}

	// 3. Loop through and move each one if it exists in the old path
	for _, fileName := range filesToMigrate {
		oldPath := filepath.Join(oldDataPath, fileName)
		newPath := filepath.Join(newDataPath, fileName)

		// Check if the old file/folder exists and the new one doesn't
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				fmt.Printf("Migrating %s to new location...\n", fileName)
				// Use os.Rename, which works for both files and directories
				if err := os.Rename(oldPath, newPath); err != nil {
					// Log the error but don't stop the app
					log.Printf("Failed to migrate %s: %v", fileName, err)
				}
			}
		}
	}
	return nil
}

func setupLogging(logPath string) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}