package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"    // Add this import
	"strings" // Add this import
	"time"

	"github.com/Harschmann/Todo-/model"
	"github.com/xuri/excelize/v2"
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

// ... (Init, SaveLog, GetAllLogs, DeleteLog, UpdateLog, and the Stats functions remain the same) ...
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
func normalizeDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
func calculateStreak(logs []model.Log) int {
	if len(logs) == 0 {
		return 0
	}
	uniqueDates := make(map[time.Time]bool)
	for _, logEntry := range logs {
		uniqueDates[normalizeDate(logEntry.Date)] = true
	}
	streak := 0
	dayToCheck := normalizeDate(time.Now())
	if !uniqueDates[dayToCheck] {
		dayToCheck = dayToCheck.AddDate(0, 0, -1)
	}
	for uniqueDates[dayToCheck] {
		streak++
		dayToCheck = dayToCheck.AddDate(0, 0, -1)
	}
	return streak
}
func GetDailyStats() (DailyStats, error) {
	var stats DailyStats
	allLogs, err := GetAllLogs()
	if err != nil {
		return stats, err
	}
	now := time.Now()
	startOfDay := normalizeDate(now)
	for _, logEntry := range allLogs {
		if logEntry.Date.After(startOfDay) {
			stats.SolvedToday++
			stats.TimeToday += logEntry.TimeSpent
		}
	}
	stats.Streak = calculateStreak(allLogs)
	return stats, nil
}

// UPDATED: This function now includes the backup rotation logic.
func BackupToJSON() error {
	logs, err := GetAllLogs()
	if err != nil {
		return fmt.Errorf("could not get logs for backup: %w", err)
	}
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

	// --- ADDED: Backup Rotation Logic ---
	const maxBackups = 10
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("could not read backup directory: %w", err)
	}

	// Filter for our backup files and sort them by name (oldest first)
	var backupFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "backup-") && strings.HasSuffix(file.Name(), ".json") {
			backupFiles = append(backupFiles, file.Name())
		}
	}
	sort.Strings(backupFiles)

	// If we have more backups than the max, delete the oldest ones
	if len(backupFiles) > maxBackups {
		filesToDelete := backupFiles[:len(backupFiles)-maxBackups]
		for _, f := range filesToDelete {
			log.Printf("Deleting old backup: %s", f)
			os.Remove(filepath.Join(backupDir, f))
		}
	}

	return nil
}

// ExportToExcel reads all logs and saves them to an Excel .xlsx file.
func ExportToExcel() (string, error) {
	logs, err := GetAllLogs()
	if err != nil {
		return "", fmt.Errorf("could not get logs for export: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Logs"
	index, _ := f.NewSheet(sheet)

	// Set header row
	headers := []string{"Date", "Platform", "Question ID", "Topic", "Difficulty", "Time Spent (mins)", "Notes"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	// Write data rows
	for i, logEntry := range logs {
		row := i + 2 // Start on the second row
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), logEntry.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), logEntry.Platform)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), logEntry.QuestionID)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), logEntry.Topic)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), logEntry.Difficulty)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), logEntry.TimeSpent)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), logEntry.Notes)
	}

	f.SetActiveSheet(index)

	// Save spreadsheet by the given path.
	fileName := "todoplusplus_logs_export.xlsx"
	if err := f.SaveAs(fileName); err != nil {
		return "", err
	}

	return fileName, nil
}
