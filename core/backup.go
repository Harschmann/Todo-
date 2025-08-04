package core

import (
	"log"
	"time"

	"github.com/Harschmann/Todo-/db"
)

// UPDATED: This function now takes the appDataDir as an argument.
func StartPeriodicBackups(appDataDir string) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	log.Println("Periodic backup service started.")

	for range ticker.C {
		log.Println("Ticker ticked. Performing backup...")
		// Pass the appDataDir to the backup function
		if err := db.BackupToJSON(appDataDir); err != nil {
			log.Printf("Failed to perform periodic backup: %v", err)
		}
	}
}
