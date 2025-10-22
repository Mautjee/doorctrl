package db

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL embed.FS

type DB struct {
	*sql.DB
}

func InitDB(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	schema, err := schemaSQL.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")

	return &DB{db}, nil
}

func (db *DB) CreateUser(username, displayName string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO users (username, display_name) VALUES (?, ?)",
		username, displayName,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetUserByUsername(username string) (int64, string, error) {
	var id int64
	var displayName string
	err := db.QueryRow(
		"SELECT id, display_name FROM users WHERE username = ?",
		username,
	).Scan(&id, &displayName)
	return id, displayName, err
}

func (db *DB) SaveCredential(userID int64, credentialID, publicKey []byte, backupEligible, backupState bool) error {
	_, err := db.Exec(
		"INSERT INTO credentials (user_id, credential_id, public_key, backup_eligible, backup_state) VALUES (?, ?, ?, ?, ?)",
		userID, credentialID, publicKey, backupEligible, backupState,
	)
	return err
}

func (db *DB) GetCredential(credentialID []byte) (int64, []byte, int, bool, bool, error) {
	var userID int64
	var publicKey []byte
	var signCount int
	var backupEligible, backupState bool
	err := db.QueryRow(
		"SELECT user_id, public_key, sign_count, backup_eligible, backup_state FROM credentials WHERE credential_id = ?",
		credentialID,
	).Scan(&userID, &publicKey, &signCount, &backupEligible, &backupState)
	return userID, publicKey, signCount, backupEligible, backupState, err
}

func (db *DB) UpdateSignCount(credentialID []byte, signCount int) error {
	_, err := db.Exec(
		"UPDATE credentials SET sign_count = ? WHERE credential_id = ?",
		signCount, credentialID,
	)
	return err
}

func (db *DB) GetCredentialsByUserID(userID int64) ([][]byte, error) {
	rows, err := db.Query(
		"SELECT credential_id FROM credentials WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials [][]byte
	for rows.Next() {
		var credID []byte
		if err := rows.Scan(&credID); err != nil {
			return nil, err
		}
		credentials = append(credentials, credID)
	}
	return credentials, nil
}

func (db *DB) CreateBooking(userID int64, startTime, endTime string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO bookings (user_id, start_time, end_time) VALUES (?, ?, ?)",
		userID, startTime, endTime,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetUserBookings(userID int64) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		"SELECT id, start_time, end_time, status, created_at FROM bookings WHERE user_id = ? ORDER BY start_time DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []map[string]interface{}
	for rows.Next() {
		var id int64
		var startTime, endTime, status, createdAt string
		if err := rows.Scan(&id, &startTime, &endTime, &status, &createdAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, map[string]interface{}{
			"id":         id,
			"start_time": startTime,
			"end_time":   endTime,
			"status":     status,
			"created_at": createdAt,
		})
	}
	return bookings, nil
}

func (db *DB) GetActiveBooking(userID int64, currentTime string) (map[string]interface{}, error) {
	var id int64
	var startTime, endTime, status string
	err := db.QueryRow(
		"SELECT id, start_time, end_time, status FROM bookings WHERE user_id = ? AND start_time <= ? AND end_time >= ? AND status = 'active' LIMIT 1",
		userID, currentTime, currentTime,
	).Scan(&id, &startTime, &endTime, &status)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         id,
		"start_time": startTime,
		"end_time":   endTime,
		"status":     status,
	}, nil
}

func (db *DB) CheckBookingConflict(userID int64, startTime, endTime string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM bookings WHERE user_id = ? AND status = 'active' AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
		userID, endTime, startTime, startTime, startTime, startTime, endTime,
	).Scan(&count)

	return count > 0, err
}
