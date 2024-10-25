package root

import (
	"net/http"
	"fmt"
	"text/template"
	database "root/backend/database" // Import the database package
	bcrypt "golang.org/x/crypto/bcrypt"
)

func ServerRunner() {

	http.HandleFunc("/", RootHandler)    // Home page
	http.HandleFunc("/auth", Auth)       // Authentication page
	http.HandleFunc("/register", Register) // Registration form
	http.HandleFunc("/login", Login)     // Login form
	fs := http.FileServer(http.Dir("./frontend/css"))
	http.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", fs))

	// Now listening for HTTPS on port 8080
	fmt.Print("The server is running on HTTPS port :8080\n")
	// err := http.ListenAndServeTLS(":8080", "./certs/cert.pem", "./certs/key.pem", nil)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
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

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username, email, password, secondPass := r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), r.FormValue("secondpass")
		if len(username) > 50 || len(email) > 50 || len(password) > 50 || len(secondPass) > 50 {
			http.Error(w, "Wrong number of values! You are allowed to enter up to 50 characters", http.StatusBadRequest)
			return
		} else if len(username) < 3 || len(email) < 8 || len(password) < 8 || len(secondPass) < 8 {
			http.Error(w, "Each input must be at least 8 characters", http.StatusBadRequest)
			return
		}
		if username == "" || email == "" || password == "" || secondPass == "" {
			http.Error(w, "All fields are required!", http.StatusBadRequest)
			return
		}
		if password != secondPass {
			http.Error(w, "Passwords do not match!", http.StatusBadRequest)
			return
		}

		// Validate username and email
		if valid, msg := ValidateInput(username, email); !valid {
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error encrypting password", http.StatusInternalServerError)
			return
		}

		// Insert user into the database
		err = database.InsertUser(email, username, string(hashedPassword))
		if err != nil {
			http.Error(w, "Error inserting user into the database", http.StatusInternalServerError)
			return
		}

		fmt.Printf("User registered:\nUsername: %v\nEmail: %s\n", username, email)
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}
}

func Login(w http.ResponseWriter, r *http .Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username, password := r.FormValue("username"), r.FormValue("password")
		if len(username) > 50 || len(password) > 50 || len(username) < 3 || len(password) < 8 {
			http.Error(w, "Username and password must be between 8-50 characters", http.StatusBadRequest)
			return
		}
		if username == "" || password == "" {
			http.Error(w, "All fields are required!", http.StatusBadRequest)
			return
		}

		// Fetch user data from the database
		email, storedHashedPassword, err := database.FetchUserByUsername(username)
		if err != nil {
			http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
			return
		}

		// Compare the provided password with the stored hashed password
		err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password))
		if err != nil {
			http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
			return
		}

		fmt.Printf("User logged in: %s (Email: %s)\n", username, email)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}