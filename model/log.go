package model

import "time"

type Log struct {
	ID              string // A unique ID for each entry (e.g. a UUID)
	QuestionID      string
	Platform        string
	Topic           string
	Difficulty      string
	TimeSpent       int
	Notes           string
	Date            time.Time
	CalendarEventID string // ADDED: To store the Google Calendar event ID
}
