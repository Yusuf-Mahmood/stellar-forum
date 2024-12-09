package root

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"root/internal/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the database connection and creates the tables if they don't exist.
func InitDB() {
	var err error
	// Initialize the global db connection
	db, err = sql.Open("sqlite3", "./internal/database/forum.db")
	if err != nil {
		log.Fatal(err)
	}

	if err2 := RunSQL(db, "./internal/database/tables.sql"); err2 != nil {
		log.Fatal(err2)
	}

	log.Println("You are connected to the database correctly")
}

func RunSQL(db *sql.DB, filepath string) error {
	sql, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	sqlComms := string(sql)
	_, err = db.Exec(sqlComms)
	return err
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

func InsertMedia(postID int64, filePath, fileType string) error {
	stmt, err := db.Prepare("INSERT INTO media (post_id, file_path, file_type) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(postID, filePath, fileType)
	return err
}

// FetchMediaByPostID retrieves all media files associated with a specific post ID
func FetchMediaByPostID(postID int) ([]models.Media, error) {
	rows, err := db.Query("SELECT file_path, file_type FROM media WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaFiles []models.Media
	for rows.Next() {
		var media models.Media
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

// LikePost adds a like for the given post and user
func LikePost(userID int, postID string) error {
	// First, check if the user has already liked or disliked this post
	var existingLikeCount int
	var existingDislikeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 1", userID, postID).Scan(&existingLikeCount)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 0", userID, postID).Scan(&existingDislikeCount)
	if err != nil {
		return err
	}

	// If the user already liked this post, do nothing
	if existingLikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 1", userID, postID)
		if err != nil {
			return err
		}
		return nil
	}

	// If the user disliked the post, remove the dislike and add a like
	if existingDislikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 0", userID, postID)
		if err != nil {
			return err
		}
	}

	// Insert a new like into the likes table
	_, err = db.Exec("INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, 1)", userID, postID)
	if err != nil {
		return err
	}

	return nil
}

// DislikePost adds a dislike for the given post and user
func DislikePost(userID int, postID string) error {
	// First, check if the user has already liked or disliked this post
	var existingLikeCount int
	var existingDislikeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 1", userID, postID).Scan(&existingLikeCount)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 0", userID, postID).Scan(&existingDislikeCount)
	if err != nil {
		return err
	}

	// If the user already disliked this post, do nothing
	if existingDislikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 0", userID, postID)
		if err != nil {
			return err
		}
		return nil
	}

	// If the user liked the post, remove the like and add a dislike
	if existingLikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 1", userID, postID)
		if err != nil {
			return err
		}
	}

	// Insert a new dislike into the likes table
	_, err = db.Exec("INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, 0)", userID, postID)
	if err != nil {
		return err
	}

	return nil
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

// InsertComment inserts a new post into the database
func InsertComment(userID int, postID int, content string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO comments (user_id, post_id, content) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(userID, postID, content)
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

// FetchUserIDBySessionToken retrieves the user ID for a given session token
func FetchUsernameByUserID(id int) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", id).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
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

// FetchPosts retrieves all posts from the database and includes like and dislike counts.
func FetchPosts() ([]models.Post, error) {
	rows, err := db.Query(`
        SELECT 
            p.id, p.user_id, u.username, p.content, p.created_at,
            COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
            COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
        GROUP BY p.id
        ORDER BY p.created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, err
		}

		post.FormatDate = FormatDate(post.CreatedAt)

		// Optionally, fetch media for each post
		media, err := FetchMediaByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Media = media

		comments, err := FetchCommentsByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comment = comments
		commentCount, err := CountComments(post.ID)
		if err != nil {
			return nil, err
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		posts = append(posts, post)
	}
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetchMemesPostsByCategoryID(categoryID int) ([]models.MemesPosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.MemesPosts
	for rows.Next() {
		var post models.MemesPosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetchGamingPostsByCategoryID(categoryID int) ([]models.GamingPosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.GamingPosts
	for rows.Next() {
		var post models.GamingPosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetcheEducationPostsByCategoryID(categoryID int) ([]models.EducationPosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.EducationPosts
	for rows.Next() {
		var post models.EducationPosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetchTechnologyPostsByCategoryID(categoryID int) ([]models.TechnologyPosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.TechnologyPosts
	for rows.Next() {
		var post models.TechnologyPosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetchSciencePostsByCategoryID(categoryID int) ([]models.SciencePosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.SciencePosts
	for rows.Next() {
		var post models.SciencePosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

// FetchMemesPostsByCategoryID retrieves posts for the Memes category by its ID
func FetchSportsPostsByCategoryID(categoryID int) ([]models.SportsPosts, error) {
	// Query to retrieve posts in the Memes category
	rows, err := db.Query(`
		SELECT p.id, pc.category_id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("Error querying posts for category ID %d: %w", categoryID, err)
	}
	defer rows.Close()

	// Initialize an empty slice to store the fetched posts
	var posts []models.SportsPosts
	for rows.Next() {
		var post models.SportsPosts

		// Scan the database row into the MemesPosts struct
		err := rows.Scan(&post.PostID, &post.CategoriesID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning post data: %w", err)
		}

		// Format the created date for the post
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media for the post (e.g., images, videos)
		media, err := FetchMediaByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching media for post %d: %w", post.PostID, err)
		}
		post.Media = media

		// Fetch comments for the post
		comments, err := FetchCommentsByPostID(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error fetching comments for post %d: %w", post.PostID, err)
		}
		post.Comment = comments

		// Fetch the comment count for the post
		commentCount, err := CountComments(post.PostID)
		if err != nil {
			return nil, fmt.Errorf("Error counting comments for post %d: %w", post.PostID, err)
		}
		post.ComCount = commentCount

		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		// Append the populated post to the posts slice
		posts = append(posts, post)
	}

	// Return the populated slice of MemesPosts
	return posts, nil
}

func FormatDate(date time.Time) string {
	return date.Format("02 Jan 2006")
}

func FetchCommentsByPostID(postID int) ([]models.Comment, error) {
	rows, err := db.Query(`
		SELECT 
			c.id, 
			c.user_id, 
			c.content, 
			c.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM 
			comments c
		LEFT JOIN 
			likes l ON c.id = l.comment_id
		WHERE 
			c.post_id = ?
		GROUP BY 
			c.id, c.user_id, c.content, c.created_at
		ORDER BY 
			c.created_at DESC;
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var userID int
		err := rows.Scan(&comment.ComID, &userID, &comment.ComContent, &comment.ComCreatedAt, &comment.ComLikes, &comment.ComDislikes)
		if err != nil {
			return nil, err
		}

		comment.ComFormatDate = FormatDate(comment.ComCreatedAt)

		// Fetch username using FetchUsernameByUserID
		username, err := FetchUsernameByUserID(userID)
		if err != nil {
			return nil, err
		}
		comment.ComUsername = username
		comment.PostID = postID
		
		profileColor, err := getProfileColor(userID)
		if err != nil {
			return nil, err
		}
		comment.ComProfile = profileColor

		comments = append(comments, comment)
	}

	return comments, nil
}

func LikeComment(userID int, postID string, commentID string) error {
	// First, check if the user has already liked or disliked this post
	var existingLikeCount int
	var existingDislikeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 1", userID, postID, commentID).Scan(&existingLikeCount)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 0", userID, postID, commentID).Scan(&existingDislikeCount)
	if err != nil {
		return err
	}

	// If the user already liked this post, do nothing
	if existingLikeCount > 0 {
		return nil
	}

	// If the user disliked the post, remove the dislike and add a like
	if existingDislikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 0", userID, postID, commentID)
		if err != nil {
			return err
		}
	}

	// Insert a new like into the likes table
	_, err = db.Exec("INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, ?, ?, 1)", userID, postID, commentID)
	if err != nil {
		return err
	}

	return nil
}

func DislikeComment(userID int, postID string, commentID string) error {
	// First, check if the user has already liked or disliked this post
	var existingLikeCount int
	var existingDislikeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 1", userID, postID, commentID).Scan(&existingLikeCount)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 0", userID, postID, commentID).Scan(&existingDislikeCount)
	if err != nil {
		return err
	}

	// If the user already liked this post, do nothing
	if existingDislikeCount > 0 {
		return nil
	}

	// If the user disliked the post, remove the dislike and add a like
	if existingLikeCount > 0 {
		_, err := db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND comment_id = ? AND is_like = 1", userID, postID, commentID)
		if err != nil {
			return err
		}
	}

	// Insert a new like into the likes table
	_, err = db.Exec("INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, ?, ?, 0)", userID, postID, commentID)
	if err != nil {
		return err
	}

	return nil
}

// CountComments fetches the total number of comments for a specific post.
func CountComments(postID int) (ComCount int, err error) {
	query := `
        SELECT 
            COUNT(*) AS ComCount
        FROM comments
        WHERE post_id = ?
    `

	err = db.QueryRow(query, postID).Scan(&ComCount)
	return
}

func FetchLikedPosts(userID int) ([]models.Post, error) {
	query := `
        SELECT 
            p.id, p.user_id, u.username, p.content, p.created_at,
            COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
            COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
        FROM likes l
        JOIN posts p ON l.post_id = p.id
        JOIN users u ON p.user_id = u.id
        WHERE l.user_id = ? AND l.is_like = 1
        GROUP BY p.id
        ORDER BY p.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likedPosts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		err := rows.Scan(&post.ID, &post.UserID, &username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, err
		}

		post.Username = username
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media and comments for the post
		post.Media, err = FetchMediaByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comment, err = FetchCommentsByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		likedPosts = append(likedPosts, post)
	}

	return likedPosts, nil
}

func FetchDislikedPosts(userID int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT 
			p.id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM likes l
		JOIN posts p ON l.post_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE l.user_id = ? AND l.is_like = 0
		GROUP BY p.id
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dislikedPosts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		err := rows.Scan(&post.ID, &post.UserID, &username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, err
		}

		post.Username = username
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media and comments for the post
		post.Media, err = FetchMediaByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comment, err = FetchCommentsByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		dislikedPosts = append(dislikedPosts, post)
	}

	return dislikedPosts, nil
}

func FetchCreatedPosts(userID int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT 
			p.id, p.user_id, u.username, p.content, p.created_at,
			COUNT(CASE WHEN l.is_like = 1 THEN 1 END) AS likes,
			COUNT(CASE WHEN l.is_like = 0 THEN 1 END) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN likes l ON p.id = l.post_id AND l.comment_id IS NULL
		WHERE p.user_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var createdPosts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		err := rows.Scan(&post.ID, &post.UserID, &username, &post.Content, &post.CreatedAt, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, err
		}

		post.Username = username
		post.FormatDate = FormatDate(post.CreatedAt)

		// Fetch media and comments for the post
		post.Media, err = FetchMediaByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comment, err = FetchCommentsByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		profileColor, err := getProfileColor(post.UserID)
		if err != nil {
			return nil, err
		}
		post.ProfileColor = profileColor

		createdPosts = append(createdPosts, post)
	}

	return createdPosts, nil
}

func FetchUserProfileBySessionToken(sessionToken string) ([]models.UserProfile, error) {

	if sessionToken == "" {
		return nil, nil
	}
	// Fetch liked, disliked, and created posts concurrently
	userID, err := FetchUserIDBySessionToken(sessionToken)
	if err != nil {
		return nil, err
	}
	likedPosts, err := FetchLikedPosts(userID)
	if err != nil {
		return nil, err
	}
	dislikedPosts, err := FetchDislikedPosts(userID)
	if err != nil {
		return nil, err
	}

	createdPosts, err := FetchCreatedPosts(userID)
	if err != nil {
		return nil, err
	}
	username, err := FetchUsernameByUserID(userID)
	if err != nil {
		return nil, err
	}
	profileColor, err := getProfileColor(userID)
	if err != nil {
		return nil, err
	}
	// Create a UserProfile struct
	userProfile := models.UserProfile{
		UserID:        userID,
		Username:      username,
		ProfileColor:  profileColor,
		LikedPosts:    likedPosts,
		DislikedPosts: dislikedPosts,
		CreatedPosts:  createdPosts,
	}

	// Wrap the UserProfile in a slice
	return []models.UserProfile{userProfile}, nil
}

func UpdateProfileColor(userID int, color string) error {
	query := `UPDATE users SET profile_color = ? WHERE id = ?`

	_, err := db.Exec(query, color, userID)
	if err != nil {
		return err
	}
	return nil
}

func getProfileColor(userID int) (string, error) {
	var color string
	query := `SELECT profile_color FROM users WHERE id = ?`

	err := db.QueryRow(query, userID).Scan(&color)
	if err != nil {
		return "", err
	}

	return color, nil
}
