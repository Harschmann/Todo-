package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Harschmann/Todo-/model"
	"go.etcd.io/bbolt"
)

// DailyStats holds the calculated statistics.
type DailyStats struct {
	SolvedToday int
	TimeToday   int
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
		b := tx.Bucket(logBucket)
		logEntry.Date = time.Now()
		key, err := logEntry.Date.MarshalText()
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(logEntry)
		if err != nil {
			return err
		}
		return b.Put(key, encoded)
	})
}

func GetAllLogs() ([]model.Log, error) {
	var logs []model.Log
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(logBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var logEntry model.Log
			if err := json.Unmarshal(v, &logEntry); err != nil {
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

func DeleteLog(date time.Time) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(logBucket)
		key, err := date.MarshalText()
		if err != nil {
			return err
		}
		return b.Delete(key)
	})
}

func UpdateLog(logEntry *model.Log) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(logBucket)
		key, err := logEntry.Date.MarshalText()
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(logEntry)
		if err != nil {
			return err
		}
		return b.Put(key, encoded)
	})
}

// NOTE: We reverted the stats features, so this can be removed or kept for later.
// To keep things simple, let's keep it but it won't be used by the TUI.
func GetTodaysStats() (DailyStats, error) {
	var stats DailyStats
	allLogs, err := GetAllLogs()
	if err != nil {
		return stats, err
	}
	now := time.Now()
	year, month, day := now.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	for _, logEntry := range allLogs {
		if logEntry.Date.After(startOfDay) {
			stats.SolvedToday++
			stats.TimeToday += logEntry.TimeSpent
		}
	}
	stats.Streak = 0 // Placeholder
	return stats, nil
}

// BackupToJSON reads all logs and writes them to a timestamped JSON file.
func BackupToJSON() error {
	logs, err := GetAllLogs()
	if err != nil {
		return fmt.Errorf("could not get logs for backup: %w", err)
	}

	// CORRECTED: Ensure we write `[]` instead of `null` for empty backups.
	if logs == nil {
		logs = []model.Log{}
	}

	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal logs to JSON: %w", err)
	}

	filename := fmt.Sprintf("backup-%s.json", time.Now().Format("2006-01-02_15-04-05"))
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("could not create backup directory: %w", err)
	}

	filePath := filepath.Join(backupDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("could not write backup file: %w", err)
	}

	log.Printf("Successfully created backup: %s", filePath)
	return nil
}
