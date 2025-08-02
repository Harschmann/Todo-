package model

import "time"

// Log represents a single competitive programming log
type Log struct {
	ID         string // A unique ID for each entry (e.g. a UUID)
	QuestionID string // ADDED: The ID from the platform (e.g., "1337A")
	Platform   string // e.g. "Codeforces", "Leetcode"
	Topic      string // e.g. "DP", "Graph", "Strings"
	Difficulty string // e.g. "Easy", "Medium", "Hard"
	TimeSpent  int    // Note: Your original struct had `Timespent`, Go convention is `TimeSpent`
	Notes      string
	Date       time.Time
}
