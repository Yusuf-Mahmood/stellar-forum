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
            email TEXT NOT NULL UNIQUE,
            username TEXT NOT NULL,
            password_hash TEXT NOT NULL,
            cookies TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
    }

    for _, query := range createTableQueries {
        _, err = db.Exec(query)
        if err != nil {
            log.Fatalf("Error creating table: %s", err)
        }
    }

    log.Println("Database and tables created successfully!")
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
func FetchUserByUsername(username string) (string, string, error) {
    var email, passwordHash string
    err := db.QueryRow("SELECT email, password_hash FROM users WHERE username = ?", username).Scan(&email, &passwordHash)
    if err != nil {
        return "", "", err
    }
    return email, passwordHash, nil
}
