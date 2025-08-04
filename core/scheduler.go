package core

import (
	"log"
	// "time"

	"github.com/Harschmann/Todo-/calendar"
	"github.com/Harschmann/Todo-/db"
)

// CheckAndSendReminder checks if a problem was solved today and sends an email if not.
// This is now a simple, single-action function.
func CheckAndSendReminder() {
	stats, err := db.GetDailyStats()
	if err != nil {
		log.Printf("Reminder check failed: could not get stats: %v", err)
		return
	}

	if stats.SolvedToday == 0 {
		log.Println("Condition met (0 pro`blems solved today). Sending daily reminder email...")
		if err := calendar.SendReminderEmail(); err != nil {
			log.Printf("Failed to send reminder email: %v", err)
		}
	} else {
		log.Printf("Condition not met (%d problems solved today). No reminder sent.", stats.SolvedToday)
	}
}