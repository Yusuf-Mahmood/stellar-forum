package root

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the database connection and creates the tables if they don't exist.
func InitDB() {
	var err error
	// Initialize the global db connection
	db, err = sql.Open("sqlite3", "./backend/database/forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	createTableQueries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT NOT NULL UNIQUE,
            username TEXT NOT NULL,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            cookies TEXT NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users (id)
        );`,
		`CREATE TABLE IF NOT EXISTS comments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES posts (id),
            FOREIGN KEY (user_id) REFERENCES users (id)
        );`,
		`CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE
        );`,
		`CREATE TABLE IF NOT EXISTS post_categories (
            post_id INTEGER NOT NULL,
            category_id INTEGER NOT NULL,
            PRIMARY KEY (post_id, category_id),
            FOREIGN KEY (post_id) REFERENCES posts (id),
            FOREIGN KEY (category_id) REFERENCES categories (id)
        );`,
		`CREATE TABLE IF NOT EXISTS likes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            post_id INTEGER NOT NULL,
            comment_id INTEGER,
            is_like BOOLEAN NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users (id),
            FOREIGN KEY (post_id) REFERENCES posts (id),
            FOREIGN KEY (comment_id) REFERENCES comments (id)
        );`,
		`CREATE TABLE IF NOT EXISTS media (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            file_path TEXT NOT NULL,
            file_type TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES posts (id)
        );`,
	}

	for _, query := range createTableQueries {
		_, err = db.Exec(query)
		if err != nil {
			log.Fatalf("Error creating table: %s", err)
		}
	}

	log.Println("You are connected to the database correctly")
}

// insertUser inserts a new user into the database
func InsertUser(email, username, passwordHash string) error {
	fmt.Println("You entered this function")
	stmt, err := db.Prepare("INSERT INTO users(email, username, password_hash, cookies) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(email, username, passwordHash, "")
	return err
}

// fetchUserByUsername fetches user data based on the username
func FetchUserByUsername(username string) (string, error) {
	var passwordHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

// StoreSessionToken saves the session token in the database
func StoreSessionToken(username, token string) error {
	_, err := db.Exec("UPDATE users SET cookies = ? WHERE username = ?", token, username)
	return err
}

// IsValidSession checks if the provided session token is valid
func IsValidSession(token string) bool {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE cookies = ?", token).Scan(&username)
	return err == nil && username != ""
}

// CheckUsernameExists checks if a user already exists with the given username
func CheckUsernameExists(username string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT COUNT(1) FROM users WHERE username = ?", username).Scan(&exists)
	return exists, err
}

// CheckEmailExists checks if a user already exists with the given email
func CheckEmailExists(email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?", email).Scan(&exists)
	return exists, err
}

func InsertMedia(postID, filePath, fileType string) error {
	stmt, err := db.Prepare("INSERT INTO media (post_id, file_path, file_type) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(postID, filePath, fileType)
	return err
}

// Media represents a media file linked to a post
type Media struct {
	FilePath string
	FileType string
}

// FetchMediaByPostID retrieves all media files associated with a specific post ID
func FetchMediaByPostID(postID int) ([]Media, error) {
	rows, err := db.Query("SELECT file_path, file_type FROM media WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaFiles []Media
	for rows.Next() {
		var media Media
		if err := rows.Scan(&media.FilePath, &media.FileType); err != nil {
			return nil, err
		}
		mediaFiles = append(mediaFiles, media)
	}
	return mediaFiles, nil
}
