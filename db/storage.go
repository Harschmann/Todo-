package db

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Harschmann/Todo-/model"
	"go.etcd.io/bbolt"
)

// DailyStats holds the calculated statistics.
type DailyStats struct {
	SolvedToday int
	TimeToday   int // in minutes
	Streak      int
}

var db *bbolt.DB
var logBucket = []byte("logs")

// Init opens the database file and creates the necessary buckets.
func Init(dbPath string) error {
	var err error
	db, err = bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}

	// Create the "logs" bucket if it doesn't already exist.
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(logBucket)
		if err != nil {
			log.Printf("could not create bucket: %v", err)
			return err
		}
		return nil
	})
}

func SaveLog(logEntry *model.Log) error {
	return db.Update(func(tx *bbolt.Tx) error {
		// 1. Get the "logs" bucket.
		b := tx.Bucket(logBucket)

		// 2. Set the log's date to the current time.
		logEntry.Date = time.Now()

		// 3. Create a key for our data. Using a timestamp ensures keys are unique and chronological.
		key, err := logEntry.Date.MarshalText()
		if err != nil {
			return err
		}

		// 4. Encode the logEntry struct into JSON bytes.
		encoded, err := json.Marshal(logEntry)
		if err != nil {
			return err
		}

		// 5. Save the key/value pair to the bucket.
		return b.Put(key, encoded)
	})
}

// GetAllLogs retrieves all log entries from the database.
func GetAllLogs() ([]model.Log, error) {
	var logs []model.Log
	err := db.View(func(tx *bbolt.Tx) error {
		// Get the logs bucket
		b := tx.Bucket(logBucket)

		// Iterate over all the key/value pairs in the bucket
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var logEntry model.Log
			// Decode the JSON value back into a Log struct
			if err := json.Unmarshal(v, &logEntry); err != nil {
				// If one entry is corrupt, we can log it and continue
				log.Printf("could not unmarshal log entry: %v", err)
				continue
			}
			logs = append(logs, logEntry)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// DeleteLog removes a log entry from the database using its date as the key.
func DeleteLog(date time.Time) error {
	return db.Update(func(tx *bbolt.Tx) error {
		// Get the logs bucket
		b := tx.Bucket(logBucket)

		// Get the key for the log we want to delete
		key, err := date.MarshalText()
		if err != nil {
			return err
		}

		// Delete the key from the bucket
		return b.Delete(key)
	})
}

// UpdateLog finds a log by its date and overwrites it with the new data.
func UpdateLog(logEntry *model.Log) error {
	return db.Update(func(tx *bbolt.Tx) error {
		// Get the logs bucket
		b := tx.Bucket(logBucket)

		// Get the key for the log we want to update.
		// It's crucial that the logEntry.Date field has not been changed.
		key, err := logEntry.Date.MarshalText()
		if err != nil {
			return err
		}

		// Marshal the updated log entry into JSON
		encoded, err := json.Marshal(logEntry)
		if err != nil {
			return err
		}

		// Save it to the bucket, overwriting the old value
		return b.Put(key, encoded)
	})
}

// GetTodaysStats calculates the number of questions solved and time spent today.
func GetTodaysStats() (DailyStats, error) {
	var stats DailyStats
	allLogs, err := GetAllLogs()
	if err != nil {
		return stats, err
	}

	// Get the start of the current day (midnight) in the local timezone.
	now := time.Now()
	year, month, day := now.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	// Loop through all logs and count the ones from today.
	for _, logEntry := range allLogs {
		if logEntry.Date.After(startOfDay) {
			stats.SolvedToday++
			stats.TimeToday += logEntry.TimeSpent
		}
	}

	// We will calculate the streak in a later step.
	stats.Streak = 0 // Placeholder

	return stats, nil
}
