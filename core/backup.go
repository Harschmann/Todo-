package core

import (
	"log"
	"time"

	"github.com/Harschmann/Todo-/db"
)

// StartPeriodicBackups starts a ticker that triggers a database backup at a set interval.
// This function should be run as a goroutine.
func StartPeriodicBackups() {
	// For testing, we'll use a short interval.
	// For production, you would change this to time.Hour.
	ticker := time.NewTicker(30 * time.Hour)
	defer ticker.Stop()

	log.Println("Periodic backup service started. A backup will be created every 30 seconds.")

	// This loop will run forever in the background.
	for range ticker.C {
		log.Println("Ticker ticked. Performing backup...")
		if err := db.BackupToJSON(); err != nil {
			// In a real app, you might want more robust error handling,
			// but for now, we'll just log the error.
			log.Printf("Failed to perform periodic backup: %v", err)
		}
	}
}
