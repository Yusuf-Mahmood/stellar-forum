package root

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	database "root/internal/database"
	"root/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func ServerRunner() {
	http.HandleFunc("/", RootHandler)                  // Home page
	http.HandleFunc("/auth", Auth)                     // Authentication page
	http.HandleFunc("/register", Register)             // Registration form
	http.HandleFunc("/login", Login)                   // Login form
	http.HandleFunc("/logout", Logout)                 // Logout
	http.HandleFunc("/createpost", CreatePost)         // Post Handler
	http.HandleFunc("/createcomment", CreateComment)   // Comment Handler
	http.HandleFunc("/auth/google", handleGoogleLogin) // Google login
	http.HandleFunc("/auth/callback", handleGoogleCallback)
	http.HandleFunc("/auth/github", handleGitHubLogin) // GitHub login
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)
	http.HandleFunc("/like", LikePost)                 // Like Handler
	http.HandleFunc("/dislike", DislikePost)           // Dislike Handler
	http.HandleFunc("/Commentlike", LikeComment)       // Comment Like Handler
	http.HandleFunc("/Commentdislike", DislikeComment) // Comment Dislike Handler
	http.HandleFunc("/inPostlike", inLikePost)
	http.HandleFunc("/inPostdislike", inDislikePost)
	http.HandleFunc("/profilePicture", UpdateProfileColor)
	http.HandleFunc("/redirect", Redirect)
	http.HandleFunc("/assets/uploads", NotFound)
	http.HandleFunc("/assets/images", NotFound)
	http.HandleFunc("/assets/static", InternalServerError)
	http.HandleFunc("/404", NotFound)
	http.HandleFunc("/500", InternalServerError)
	http.HandleFunc("/400", BadRequest)
	http.HandleFunc("/405", Mnotallowed)
	fs := http.FileServer(http.Dir("./assets/static"))
	http.Handle("/assets/static/", http.StripPrefix("/assets/static/", fs))

	fs2 := http.FileServer(http.Dir("./assets/images"))
	http.Handle("/assets/images/", http.StripPrefix("/assets/images/", fs2))

	fs3 := http.FileServer(http.Dir("./assets/uploads"))
	http.Handle("/assets/uploads/", http.StripPrefix("/assets/uploads/", fs3))

	fmt.Print("The server is running on https://localhost:8080/\n")
	err := http.ListenAndServeTLS(":8080", "./internal/certs/cert.pem", "./internal/certs/key.pem", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

var (
	maxPosts = 0
)

func Redirect(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/#post=%s", postID), http.StatusSeeOther)
}

// RootHandler checks if a user is logged in and redirects accordingly
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	if r.URL.Path != "/" {
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	isGuest := false
	var userID int

	// Check session token
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken == nil || sessionToken.Value == "" {
		isGuest = true
	} else {
		userID, err = database.FetchUserIDBySessionToken(sessionToken.Value)
		if err != nil {
			isGuest = true
		}
	}

	// Initialize post variables
	var posts []models.Post
	var memesPosts []models.MemesPosts
	var gamingPosts []models.GamingPosts
	var educationPosts []models.EducationPosts
	var technologyPosts []models.TechnologyPosts
	var sciencePosts []models.SciencePosts
	var sportsPosts []models.SportsPosts

	// Fetch posts and categories for both guests and logged-in users
	posts, err = database.FetchPosts(userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Fetch category posts
	memesPosts, err = database.FetchMemesPostsByCategoryID(2, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	gamingPosts, err = database.FetchGamingPostsByCategoryID(3, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	educationPosts, err = database.FetcheEducationPostsByCategoryID(4, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	technologyPosts, err = database.FetchTechnologyPostsByCategoryID(5, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	sciencePosts, err = database.FetchSciencePostsByCategoryID(6, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	sportsPosts, err = database.FetchSportsPostsByCategoryID(7, userID)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	if isGuest {
		// Load guest template
		t, terr := template.ParseFiles("./assets/templates/guesthome.html")
		if terr != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}

		guestData := struct {
			Post            []models.Post
			MemesPosts      []models.MemesPosts
			GamingPosts     []models.GamingPosts
			EducationPosts  []models.EducationPosts
			TechnologyPosts []models.TechnologyPosts
			SciencePosts    []models.SciencePosts
			SportsPosts     []models.SportsPosts
		}{
			Post:            posts,
			MemesPosts:      memesPosts,
			GamingPosts:     gamingPosts,
			EducationPosts:  educationPosts,
			TechnologyPosts: technologyPosts,
			SciencePosts:    sciencePosts,
			SportsPosts:     sportsPosts,
		}

		err = t.Execute(w, guestData)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
		}
		return
	}

	// Load logged-in user template and continue with user-specific logic
	t, terr := template.ParseFiles("./assets/templates/home.html")
	if terr != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Fetch user profile
	userProfile, err := database.FetchUserProfileBySessionToken(sessionToken.Value)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Prepare data for the template
	data := models.Data{
		UserProfile:     userProfile,
		Post:            posts,
		MemesPosts:      memesPosts,
		GamingPosts:     gamingPosts,
		EducationPosts:  educationPosts,
		TechnologyPosts: technologyPosts,
		SciencePosts:    sciencePosts,
		SportsPosts:     sportsPosts,
	}

	// Execute template with user data
	err = t.Execute(w, data)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
}

// Auth handles the authentication page
func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	if r.URL.Path != "/auth" {
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	t, err := template.ParseFiles("./assets/templates/auth.html")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	err2 := t.Execute(w, nil)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/400", http.StatusSeeOther)
			return
		}

		username, password, secondpass, email := strings.TrimSpace(r.FormValue("username")), r.FormValue("password"), r.FormValue("secondpass"), r.FormValue("email")
		// Validate input lengths
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			renderRegisterPage(w, r, "Username must be between 3-5 character and password must be between 8-50 character")
			return
		}
		if secondpass != password {
			renderRegisterPage(w, r, "Passwords do not match")
			return
		}
		if username == "" || password == "" || email == "" {
			renderRegisterPage(w, r, "All fields are required!")
			return
		}

		valid, msg := ValidateInput(username, email)
		if !valid {
			renderRegisterPage(w, r, msg)
			return
		}

		// Check if the username already exists
		exists, err := database.CheckUsernameExists(username)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		if exists {
			renderRegisterPage(w, r, "Username already taken")
			return
		}

		// Check if the username already exists
		existsEmail, err := database.CheckEmailExists(email)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		if existsEmail {
			renderRegisterPage(w, r, "Email already taken")
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}

		// Save the new user in the database using the existing InsertUser function
		err = database.InsertUser(email, username, string(hashedPassword))
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}

		// Redirect to login after successful registration
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/auth#container2", http.StatusSeeOther)
	}
}

var templates = template.Must(template.ParseGlob("./assets/templates/*.html"))

// Login handles user login and renders the login page with error messages if needed
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			renderLoginPage(w, r, "Invalid form data")
			return
		}

		username, password := r.FormValue("username"), r.FormValue("password")
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			renderLoginPage(w, r, "Username must be between 3-50 characters and password between 8-50 characters")
			return
		}
		if username == "" || password == "" {
			renderLoginPage(w, r, "All fields are required")
			return
		}

		storedHashedPassword, err := database.FetchUserByUsername(username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password)) != nil {
			renderLoginPage(w, r, "Invalid username or password")
			return
		}

		sessionToken, err := uuid.NewV4()
		if err != nil {
			renderLoginPage(w, r, "Error creating session")
			return
		}

		err = database.StoreSessionToken(username, sessionToken.String())
		if err != nil {
			renderLoginPage(w, r, "Error storing session")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken.String(),
			Expires:  time.Now().Add(1 * time.Hour),
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		renderLoginPage(w, r, "")
	}
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Check if the user has a session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// Redirect to login if no session is found
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	// Remove the session from the database
	err = database.DeleteSession(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // Expire immediately
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect to the login page after logout
	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

// renderLoginPage renders the login page with an optional error message
func renderLoginPage(w http.ResponseWriter, r *http.Request, errorMessage string) {
	data := struct {
		ErrorMessage    string
		RegErrorMessage string
	}{
		ErrorMessage:    errorMessage,
		RegErrorMessage: "",
	}

	err := templates.ExecuteTemplate(w, "auth.html", data)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}
}

// renderLoginPage renders the login page with an optional error message
func renderRegisterPage(w http.ResponseWriter, r *http.Request, errorMessage string) {
	data := struct {
		ErrorMessage    string
		RegErrorMessage string
	}{
		RegErrorMessage: errorMessage,
		ErrorMessage:    "",
	}

	err := templates.ExecuteTemplate(w, "authreg.html", data)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}
}

// NotFound handles 404 errors
func NotFound(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./assets/templates/errors/404.html")
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
}

// BadRequest handles 400 errors
func BadRequest(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./assets/templates/errors/400.html")
	if err != nil {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}
}

// Mnotallowed handles 405 errors
func Mnotallowed(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./assets/templates/errors/405.html")
	if err != nil {
		http.Error(w, "405 method not allowed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "405 method not allowed", http.StatusBadRequest)
		return
	}
}

// InternalServerError handles 500 errors
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./assets/templates/errors/500.html")
	if err != nil {
		http.Error(w, "500 not found", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "500 not found", http.StatusInternalServerError)
		return
	}
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("postText"))
	if content == "" || len(content) > 366 {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	categories := r.Form["catInputs"] // Extract selected categories

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := database.InsertPost(userID, content)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	if len(categories) > 0 {
		for _, category := range categories {
			categoryID, err := database.GetOrCreateCategory(category)
			if err != nil {
				http.Redirect(w, r, "/500", http.StatusSeeOther)
				return
			}
			err = database.AssociatePostWithCategory(postID, categoryID)
			if err != nil {
				http.Redirect(w, r, "/500", http.StatusSeeOther)
				return
			}
		}
	} else {
		err = database.AssociatePostWithCategory(postID, 1)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
	}
	// Handle media upload if present in the form
	mediaFile, fileHeader, err := r.FormFile("postImage")
	if err == nil {
		if fileHeader.Size >= 20*1024*1024 {
			http.Redirect(w, r, "/400", http.StatusSeeOther)
			return
		}
		// Create the uploads directory if it doesn't exist
		uploadDir := "./assets/uploads"
		os.MkdirAll(uploadDir, os.ModePerm)

		fileExtension := filepath.Ext(fileHeader.Filename)
		if fileExtension == "" || fileExtension == ".mp4" || fileExtension == ".mov" || fileExtension == ".avi" || fileExtension != ".jpg" && fileExtension != ".png" && fileExtension != ".jpeg" && fileExtension != ".svg" && fileExtension != ".gif" {
			http.Redirect(w, r, "/400", http.StatusSeeOther)
			return
		}
		// Create a unique file name and save the file
		fileName := fmt.Sprintf("%d-%s%s", time.Now().Unix(), "postImage", fileExtension)
		filePath := filepath.Join(uploadDir, fileName)
		filePath = filepath.ToSlash(filePath)

		// Save the uploaded file
		dst, err := os.Create(filePath)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, mediaFile)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}

		// Determine file type (image or video) based on the file extension
		fileType := "image"
		ext := filepath.Ext(r.FormValue("postImage"))
		if ext == ".mp4" || ext == ".mov" || ext == ".avi" {
			fileType = "video"
		}

		// Save media details in the database
		err = database.InsertMedia(postID, filePath, fileType)
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
	}

	maxPosts++
	// Redirect to the homepage after successful post creation
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("commentInput"))
	postID := strings.TrimSpace(r.FormValue("hiddenID"))

	intPostID, err := strconv.Atoi(postID)
	if err != nil {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}
	sessionToken, _ := r.Cookie("session_token")
	userID, err := database.FetchUserIDBySessionToken(sessionToken.Value)
	if err != nil {
		return
	}
	posts, err := database.FetchPosts(userID)
	if err != nil {
		http.Redirect(w, r, "/400", http.StatusFound)
		return
	}
	for i, post := range posts {
		if post.ID == intPostID {
			break
		} else if i == len(posts)-1 {
			http.Redirect(w, r, "/400", http.StatusSeeOther)
			return
		}
	}

	if content == "" || len(content) > 366 {
		http.Redirect(w, r, "/400", http.StatusFound)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err = database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = database.InsertComment(userID, intPostID, content)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Redirect to the homepage after successful post creation
	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

// LikePost handles like action
func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.LikePost(userID, postID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#post=%s", postID), http.StatusSeeOther)
}

// DislikePost handles dislike action
func DislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.DislikePost(userID, postID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#post=%s", postID), http.StatusSeeOther)
}

func inLikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.LikePost(userID, postID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

func inDislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.DislikePost(userID, postID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

func LikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	CommentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")
	if CommentID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.LikeComment(userID, postID, CommentID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("#CommentSection=%s", postID), http.StatusSeeOther)
}

func DislikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	CommentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")
	if CommentID == "" {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Check if user has already liked or disliked the post
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err2 := database.DislikeComment(userID, postID, CommentID)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("#CommentSection=%s", postID), http.StatusSeeOther)
}

func UpdateProfileColor(w http.ResponseWriter, r *http.Request) {
	// Allow only POST requests
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/405", http.StatusSeeOther)
		return
	}

	// Extract the selected color from the form
	profileColor := r.FormValue("profileColor")
	if profileColor == "" || !isValidColor(profileColor) {
		http.Redirect(w, r, "/400", http.StatusSeeOther)
		return
	}

	// Retrieve the session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusUnauthorized)
		return
	}

	// Fetch the user ID using the session token
	userID, err := database.FetchUserIDBySessionToken(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update the profile color in the database
	err = database.UpdateProfileColor(userID, profileColor)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Redirect or respond with success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
