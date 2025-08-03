package core

import (
	"log"
	"time"

	"github.com/Harschmann/Todo-/calendar"
	"github.com/Harschmann/Todo-/db"
)

var lastReminderSent time.Time

// StartDailyReminder checks periodically if a reminder should be sent.
func StartDailyReminder() {
	// For testing, the ticker checks every minute.
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Daily reminder service started.")

	checkAndSendReminder()
	for range ticker.C {
		checkAndSendReminder()
	}
}

func checkAndSendReminder() {
	now := time.Now()

	// UPDATED: Temporarily set to 2 AM for testing.
	// Change this back to your preferred hour (e.g., 20 for 8 PM) for the final version.
	if now.Hour() != 2 {
		return
	}

	if normalizeDate(now).Equal(normalizeDate(lastReminderSent)) {
		return
	}

	stats, err := db.GetDailyStats()
	if err != nil {
		log.Printf("Reminder check failed: could not get stats: %v", err)
		return
	}

	if stats.SolvedToday == 0 {
		log.Println("Condition met. Sending daily reminder email...")
		if err := calendar.SendReminderEmail(); err != nil {
			log.Printf("Failed to send reminder email: %v", err)
		}
		lastReminderSent = now
	}
}

// normalizeDate returns the given time at the beginning of the day (00:00:00).
func normalizeDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
