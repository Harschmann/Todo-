package model

import "time"

// Log represents a single competitive programming log
type Log struct {
	ID          string       // A unique ID for each entry (e.g. a UUID)
	Platform    string       // e.g. "Codeforces", "Leetcode"
	Topic 		string       // e.g. "DP", "Graph", "Strings"
	Difficulty  string       // e.g. "Easy", "Medium", "Hard"
	Timespent   int          // Time spent in minutes
	Notes       string
	Date        time.Time
}