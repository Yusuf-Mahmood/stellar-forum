package root

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
	"database/sql"
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
            username TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL,
            cookies TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`, 
            // will write the rest of tables later ono
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
	fmt.Println("You enterd this function")
    stmt, err := db.Prepare("INSERT INTO users(email, username, password_hash) VALUES(?, ?, ?)")

    if err != nil {
        return err
    }
	
    _, err = stmt.Exec(email, username, passwordHash)
    return err
}
	
// fetchUserByUsername fetches user data based on the username
func FetchUserByUsername(username string) ( string, error) {
    var  passwordHash string
    err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&passwordHash)
    if err != nil {
        return  "", err
    }
    return  passwordHash, nil
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
