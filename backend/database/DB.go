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
            cookies TEXT
        );`,

		`CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
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

// FetchUserByUsername fetches user data based on the username
func FetchUserByUsername(username string) (string, error) {
	var userID int
	var passwordHash string
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func StoreSessionToken(username string, token string) error {
	// Prepare the SQL query to update the user's session token in the users table
	stmt, err := db.Prepare("UPDATE users SET cookies = ? WHERE username = ?")
	if err != nil {
		return err
	}
	// Execute the prepared statement with the session token and username
	_, err = stmt.Exec(token, username)
	return err
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

// DeleteSession removes a session token from the sessions table
func DeleteSession(sessionToken string) error {
	// Parameterized query to safely remove the session token
	query := "UPDATE users SET cookies = NULL WHERE cookies = ?"

	// Execute the query with the provided session token
	_, err := db.Exec(query, sessionToken)

	// Return any error encountered
	return err
}

// CreatePostWithMedia inserts a new post along with its associated media files into the database.
func CreatePostWithMedia(userID int, content string, mediaFiles []Media) (int, error) {
	// Start a database transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	// Ensure transaction is either committed or rolled back
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw the panic after rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert the post into the `posts` table
	postStmt, err := tx.Prepare("INSERT INTO posts (user_id, content) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer postStmt.Close()

	// Execute the post insertion
	result, err := postStmt.Exec(userID, content)
	if err != nil {
		return 0, err
	}

	// Retrieve the newly created post ID
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert associated media files, if any
	if len(mediaFiles) > 0 {
		mediaStmt, err := tx.Prepare("INSERT INTO media (post_id, file_path, file_type) VALUES (?, ?, ?)")
		if err != nil {
			return 0, err
		}
		defer mediaStmt.Close()

		for _, media := range mediaFiles {
			_, err = mediaStmt.Exec(postID, media.FilePath, media.FileType)
			if err != nil {
				return 0, err
			}
		}
	}

	return int(postID), nil
}

// AddOrUpdateLike adds a like or updates an existing one for a post or comment.
func AddOrUpdateLike(userID, postID int, commentID *int, isLike bool) error {
	// Check if the like already exists
	var existingID int
	query := "SELECT id FROM likes WHERE user_id = ? AND post_id = ? AND (comment_id = ? OR (comment_id IS NULL AND ? IS NULL))"
	err := db.QueryRow(query, userID, postID, commentID, commentID).Scan(&existingID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Insert a new like if none exists
			stmt, err := db.Prepare("INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, ?, ?, ?)")
			if err != nil {
				return err
			}
			defer stmt.Close()

			_, err = stmt.Exec(userID, postID, commentID, isLike)
			return err
		}
		return err
	}

	// Update the existing like
	stmt, err := db.Prepare("UPDATE likes SET is_like = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isLike, existingID)
	return err
}

// CountLikes fetches the total likes and dislikes for a specific post or comment.
func CountLikes(postID int, commentID *int) (likes int, dislikes int, err error) {
	query := `
        SELECT 
            SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) AS likes,
            SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) AS dislikes
        FROM likes
        WHERE post_id = ? AND (comment_id = ? OR (comment_id IS NULL AND ? IS NULL))
    `

	err = db.QueryRow(query, postID, commentID, commentID).Scan(&likes, &dislikes)
	return
}

// FetchUserLikes retrieves all likes made by a user.
func FetchUserLikes(userID int) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
        SELECT 
            l.id AS like_id, l.post_id, l.comment_id, l.is_like, p.content AS post_content, c.content AS comment_content
        FROM likes l
        LEFT JOIN posts p ON l.post_id = p.id
        LEFT JOIN comments c ON l.comment_id = c.id
        WHERE l.user_id = ?
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []map[string]interface{}
	for rows.Next() {
		var likeID, postID, commentID sql.NullInt64
		var isLike bool
		var postContent, commentContent sql.NullString

		err = rows.Scan(&likeID, &postID, &commentID, &isLike, &postContent, &commentContent)
		if err != nil {
			return nil, err
		}

		like := map[string]interface{}{
			"like_id":         likeID.Int64,
			"post_id":         postID.Int64,
			"comment_id":      commentID.Int64,
			"is_like":         isLike,
			"post_content":    postContent.String,
			"comment_content": commentContent.String,
		}
		likes = append(likes, like)
	}
	return likes, nil
}
