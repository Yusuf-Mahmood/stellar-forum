package root

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	database "root/backend/database"
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
	http.HandleFunc("/auth/google", handleGoogleLogin) // Google login
	http.HandleFunc("/auth/callback", handleGoogleCallback)
	http.HandleFunc("/auth/github", handleGitHubLogin) // GitHub login
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)
	http.HandleFunc("/404", NotFound)
	http.HandleFunc("/500", InternalServerError)
	fs := http.FileServer(http.Dir("./frontend/css"))
	http.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", fs))

	fmt.Print("The server is running on HTTPS port :8080\n")
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

	// Check for session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil || !database.IsValidSession(cookie.Value) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	// Render homepage if session is valid
	t, err := template.ParseFiles("./frontend/html/home.html")
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

// Auth handles the authentication page
func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

		username, password, email := strings.TrimSpace(r.FormValue("username")), r.FormValue("password"), r.FormValue("email")
		// Validate input lengths
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			renderRegisterPage(w, "Username must be between 3-5 character and password must be between 8-50 character")
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
		ErrorMessage string
	}{
		ErrorMessage: errorMessage,
	}

	err := templates.ExecuteTemplate(w, "auth.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// renderLoginPage renders the login page with an optional error message
func renderRegisterPage(w http.ResponseWriter, errorMessage string) {
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMessage,
	}

	err := templates.ExecuteTemplate(w, "auth.html", data)
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

func UploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data with a max upload size of 10 MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Retrieve the file from form data
	file, handler, err := r.FormFile("media")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the uploads directory if it doesn't exist
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	// Create a unique file name and save the file
	fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), handler.Filename)
	filePath := filepath.Join(uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Determine file type (image or video) based on file extension
	fileType := "image"
	ext := filepath.Ext(handler.Filename)
	if ext == ".mp4" || ext == ".mov" || ext == ".avi" {
		fileType = "video"
	}

	// Save media file details in the database
	postID := r.FormValue("post_id") // Ensure post_id is provided in the form data
	err = database.InsertMedia(postID, filePath, fileType)
	if err != nil {
		http.Error(w, "Error saving media details", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully")
}
