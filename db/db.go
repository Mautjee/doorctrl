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

func (db *DB) SaveCredential(userID int64, credentialID, publicKey []byte) error {
	_, err := db.Exec(
		"INSERT INTO credentials (user_id, credential_id, public_key) VALUES (?, ?, ?)",
		userID, credentialID, publicKey,
	)
	return err
}

func (db *DB) GetCredential(credentialID []byte) (int64, []byte, int, error) {
	var userID int64
	var publicKey []byte
	var signCount int
	err := db.QueryRow(
		"SELECT user_id, public_key, sign_count FROM credentials WHERE credential_id = ?",
		credentialID,
	).Scan(&userID, &publicKey, &signCount)
	return userID, publicKey, signCount, err
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
