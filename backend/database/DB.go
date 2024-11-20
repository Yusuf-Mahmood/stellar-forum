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

		`CREATE TABLE post_categories (
    	post_id INT NOT NULL,
    	category_id INT NOT NULL,
    	PRIMARY KEY (post_id, category_id),
    	FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
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
		`INSERT INTO categories (name) VALUES ('Gnrl'), ('Memes'), ('Gaming'), ('Education'), ('Technology'), ('Science'), ('Sports');`,
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

// InsertPost inserts a new post into the database
func InsertPost(userID int, content string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO posts (user_id, content) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(userID, content)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FetchUserIDBySessionToken retrieves the user ID for a given session token
func FetchUserIDBySessionToken(sessionToken string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE cookies = ?", sessionToken).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// GetOrCreateCategory returns the category ID for a given name, creating it if necessary.
func GetOrCreateCategory(name string) (int, error) {
	var categoryID int
	err := db.QueryRow("SELECT id FROM categories WHERE name = ?", name).Scan(&categoryID)
	if err == sql.ErrNoRows {
		gnrl := "Gnrl"
		err := db.QueryRow("SELECT id FROM categories WHERE name = ?", gnrl).Scan(&categoryID)
		if err != nil {
			return 0, err
		}
		return categoryID, err
	}
	return categoryID, err
}

// AssociatePostWithCategory creates an association between a post and a category.
func AssociatePostWithCategory(postID int64, categoryID int) error {
	_, err := db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
	return err
}
