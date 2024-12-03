package root

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	database "root/backend/database"
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
	http.HandleFunc("/uploads", NotFound)
	http.HandleFunc("/images", NotFound)
	http.HandleFunc("/frontend/css", InternalServerError)
	http.HandleFunc("/404", NotFound)
	http.HandleFunc("/500", InternalServerError)
	fs := http.FileServer(http.Dir("./frontend/css"))
	http.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", fs))

	fs2 := http.FileServer(http.Dir("./images"))
	http.Handle("/images/", http.StripPrefix("/images/", fs2))

	fs3 := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs3))

	fmt.Print("The server is running on https://localhost:8080/\n")
	err := http.ListenAndServeTLS(":8080", "./certs/cert.pem", "./certs/key.pem", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

// RootHandler checks if a user is logged in and redirects accordingly
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	// Render homepage if session is valid
	t, terr := template.ParseFiles("./frontend/html/home.html")
	if terr != nil {
		fmt.Println("Here4")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	posts, err := database.FetchPosts()
	if err != nil {
		fmt.Println("Here2")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" {
		// Redirect to Guest homepage if no session is found
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}

	userProfile, err := database.FetchUserProfileBySessionToken(sessionToken.Value)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}

	memesPosts, err := database.FetchMemesPostsByCategoryID(2)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}

	gamingPosts, err := database.FetchGamingPostsByCategoryID(3)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}
	educationPosts, err := database.FetcheEducationPostsByCategoryID(4)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}
	technologyPosts, err := database.FetchTechnologyPostsByCategoryID(5)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}
	sciencePosts, err := database.FetchSciencePostsByCategoryID(6)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}
	sportsPosts, err := database.FetchSportsPostsByCategoryID(7)
	if err != nil {
		t, terr = template.ParseFiles("./frontend/html/guesthome.html")
		if terr != nil {
			fmt.Println("Here3")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err = t.Execute(w, posts)
		if err != nil {
			fmt.Println("Here6")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		return
	}
	
	type Data struct {
		UserProfile []database.UserProfile
		Post       []database.Post
		MemesPosts []database.MemesPosts
		GamingPosts []database.GamingPosts
		EducationPosts []database.EducationPosts
		TechnologyPosts []database.TechnologyPosts
		SciencePosts []database.SciencePosts
		SportsPosts []database.SportsPosts
	}
	// Prepare data for the template
	data := Data{
		UserProfile: userProfile,
		Post:       posts,
		MemesPosts: memesPosts,
		GamingPosts: gamingPosts,
		EducationPosts: educationPosts,
		TechnologyPosts: technologyPosts,
		SciencePosts: sciencePosts,
		SportsPosts: sportsPosts,
	}

	// Pass posts data with like/dislike functionality to the template
	err2 := t.Execute(w, data)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
}

// Auth handles the authentication page
func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/auth" {
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	t, err := template.ParseFiles("./frontend/html/auth.html")
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username, password, secondpass, email := strings.TrimSpace(r.FormValue("username")), r.FormValue("password"), r.FormValue("secondpass"), r.FormValue("email")
		// Validate input lengths
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			renderRegisterPage(w, "Username must be between 3-5 character and password must be between 8-50 character")
			return
		}
		if secondpass != password {
			renderRegisterPage(w, "Passwords do not match")
			return
		}
		if username == "" || password == "" || email == "" {
			renderRegisterPage(w, "All fields are required!")
			return
		}

		valid, msg := ValidateInput(username, email)
		if !valid {
			renderRegisterPage(w, msg)
			return
		}

		// Check if the username already exists
		exists, err := database.CheckUsernameExists(username)
		if err != nil {
			http.Error(w, "Error checking username availability", http.StatusInternalServerError)
			return
		}
		if exists {
			renderRegisterPage(w, "Username already taken")
			return
		}

		// Check if the username already exists
		existsEmail, err := database.CheckEmailExists(email)
		if err != nil {
			http.Error(w, "Error checking Email availability", http.StatusInternalServerError)
			return
		}
		if existsEmail {
			renderRegisterPage(w, "Email already taken")
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Save the new user in the database using the existing InsertUser function
		err = database.InsertUser(email, username, string(hashedPassword))
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}

		// Redirect to login after successful registration
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/auth#container2", http.StatusSeeOther)
	}
}

var templates = template.Must(template.ParseGlob("./frontend/html/*.html"))

// Login handles user login and renders the login page with error messages if needed
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			renderLoginPage(w, "Invalid form data")
			return
		}

		username, password := r.FormValue("username"), r.FormValue("password")
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			renderLoginPage(w, "Username must be between 3-50 characters and password between 8-50 characters")
			return
		}
		if username == "" || password == "" {
			renderLoginPage(w, "All fields are required")
			return
		}

		storedHashedPassword, err := database.FetchUserByUsername(username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password)) != nil {
			renderLoginPage(w, "Invalid username or password")
			return
		}

		sessionToken, err := uuid.NewV4()
		if err != nil {
			renderLoginPage(w, "Error creating session")
			return
		}

		err = database.StoreSessionToken(username, sessionToken.String())
		if err != nil {
			renderLoginPage(w, "Error storing session")
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
		renderLoginPage(w, "")
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
		http.Error(w, "Error logging out", http.StatusInternalServerError)
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
func renderLoginPage(w http.ResponseWriter, errorMessage string) {
	data := struct {
		ErrorMessage    string
		RegErrorMessage string
	}{
		ErrorMessage:    errorMessage,
		RegErrorMessage: "",
	}

	err := templates.ExecuteTemplate(w, "auth.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// renderLoginPage renders the login page with an optional error message
func renderRegisterPage(w http.ResponseWriter, errorMessage string) {
	data := struct {
		ErrorMessage    string
		RegErrorMessage string
	}{
		RegErrorMessage: errorMessage,
		ErrorMessage:    "",
	}

	err := templates.ExecuteTemplate(w, "authreg.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// NotFound handles 404 errors
func NotFound(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./frontend/html/errors/404.html")
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

// InternalServerError handles 500 errors
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./frontend/html/errors/500.html")
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := strings.TrimSpace(r.FormValue("postText"))
	if content == "" || len(content) > 366 {
		http.Error(w, "Post content cannot be empty or exceeded limits", http.StatusBadRequest)
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
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	if len(categories) > 0 {
		for _, category := range categories {
			categoryID, err := database.GetOrCreateCategory(category)
			if err != nil {
				http.Error(w, "Failed to process category: "+category, http.StatusInternalServerError)
				return
			}
			err = database.AssociatePostWithCategory(postID, categoryID)
			if err != nil {
				http.Error(w, "Failed to associate category: "+category, http.StatusInternalServerError)
				return
			}
		}
	} else {
		err = database.AssociatePostWithCategory(postID, 1)
		if err != nil {
			http.Error(w, "Failed to associate category: General", http.StatusInternalServerError)
			return
		}
	}
	// Handle media upload if present in the form
	mediaFile, fileHeader, err := r.FormFile("postImage")
	if err == nil {
		// Create the uploads directory if it doesn't exist
		uploadDir := "./uploads"
		os.MkdirAll(uploadDir, os.ModePerm)

		fileExtension := filepath.Ext(fileHeader.Filename)
		if fileExtension == "" {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}
		// Create a unique file name and save the file
		fileName := fmt.Sprintf("%d-%s%s", time.Now().Unix(), "postImage", fileExtension)
		filePath := filepath.Join(uploadDir, fileName)
		filePath = filepath.ToSlash(filePath)

		// Save the uploaded file
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save media file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, mediaFile)
		if err != nil {
			http.Error(w, "Failed to save media file", http.StatusInternalServerError)
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
			http.Error(w, "Error saving media details", http.StatusInternalServerError)
			return
		}
	}
	// Redirect to the homepage after successful post creation
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := strings.TrimSpace(r.FormValue("commentInput"))
	postID := strings.TrimSpace(r.FormValue("hiddenID"))
	intPostID, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if content == "" || len(content) > 366 {
		http.Error(w, "Comment content cannot be empty or exceeded limits", http.StatusBadRequest)
		return
	}

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

	_, err = database.InsertComment(userID, intPostID, content)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Redirect to the homepage after successful post creation
	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

// LikePost handles like action
func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#post=%s", postID), http.StatusSeeOther)
}

// DislikePost handles dislike action
func DislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#post=%s", postID), http.StatusSeeOther)
}

func inLikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

func inDislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/#CommentSection=%s", postID), http.StatusSeeOther)
}

func LikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	CommentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")
	if CommentID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("#CommentSection=%s", postID), http.StatusSeeOther)
}

func DislikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	CommentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")
	if CommentID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("#CommentSection=%s", postID), http.StatusSeeOther)
}
