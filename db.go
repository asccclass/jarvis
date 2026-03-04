package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// InitDB initializes the SQLite database at the specified DBPath.
// If DBPath is a directory, it creates 'offline_messages.db' inside it.
func InitDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data"
	}

	// Check if the path is a directory
	info, err := os.Stat(dbPath)
	if err == nil && info.IsDir() {
		dbPath = filepath.Join(dbPath, "offline_messages.db")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS offline_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Printf("✅ Offline Database initialized at: %s\n", dbPath)
	return db, nil
}

// SaveOfflineMessage saves a message meant for a user into the DB
func SaveOfflineMessage(db *sql.DB, userID string, message string) error {
	insertSQL := `INSERT INTO offline_messages (user_id, message, created_at) VALUES (?, ?, ?)`
	_, err := db.Exec(insertSQL, userID, message, time.Now())
	if err != nil {
		return fmt.Errorf("could not save offline message: %v", err)
	}
	fmt.Printf("📦 Saved offline message for [%s]\n", userID)
	return nil
}

// GetOfflineMessages retrieves all pending messages for a specific user
func GetOfflineMessages(db *sql.DB, userID string) ([]string, error) {
	querySQL := `SELECT message FROM offline_messages WHERE user_id = ? ORDER BY created_at ASC`
	rows, err := db.Query(querySQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// DeleteOfflineMessages clears all retrieved messages for a specific user
func DeleteOfflineMessages(db *sql.DB, userID string) error {
	deleteSQL := `DELETE FROM offline_messages WHERE user_id = ?`
	_, err := db.Exec(deleteSQL, userID)
	return err
}
