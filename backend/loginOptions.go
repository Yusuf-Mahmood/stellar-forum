package root

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	database "root/backend/database"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

var (
	clientID        = "216542017821-1n8c6f6qvllcbcis3g4ohjl99imgju6r.apps.googleusercontent.com"
	clientSecret    = "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"
	redirectURI     = "https://localhost:8080/auth/callback"
	gitclientID     = "Ov23liPieKMThOkBnEuc"
	gitclientSecret = "71e4c478f2f305fada52f6187a143898d959db11"
	gitRedirectURI  = "https://localhost:8080/auth/github/callback"
)

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile&state=%s",
		url.QueryEscape(clientID), url.QueryEscape(redirectURI), "random")
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify the state parameter to protect against CSRF attacks
	if r.URL.Query().Get("state") != "random" {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}

	// Exchange the code for a token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		http.Error(w, "Failed to request token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	// Parse the JSON response to get the access token
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		http.Error(w, "Access token not found", http.StatusInternalServerError)
		return
	}

	setGoogleUserInfo(w, r, accessToken)
}

func setGoogleUserInfo(w http.ResponseWriter, r *http.Request, accessToken string) {
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(accessToken)
	resp, err := http.Get(userInfoURL)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	email := userInfo["email"].(string)
	username := userInfo["name"].(string)

	// Check if user already exists in the database
	exists, err := database.CheckEmailExists(email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		err = database.InsertUser(email, username, "")
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}
	}

	/// Create a unique session token using gofrs/uuid
	sessionToken, err := uuid.NewV4() // This generates a new UUID version 4
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Convert UUID to string for storage
	err = database.StoreSessionToken(username, sessionToken.String())
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Expires:  time.Now().Add(1 * time.Hour), // 1 hour only
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		url.QueryEscape(gitclientID),
		url.QueryEscape(gitRedirectURI),
		"random",
	)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != "random" {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}

	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", gitclientID)
	data.Set("client_secret", gitclientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", gitRedirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		http.Error(w, "Access token not found", http.StatusInternalServerError)
		return
	}

	// Fetch GitHub user info
	setGitHubUserInfo(w, r, accessToken)
}

func setGitHubUserInfo(w http.ResponseWriter, r *http.Request, accessToken string) {
	userInfoURL := "https://api.github.com/user"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		http.Error(w, "Failed to create user info request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	username, ok := userInfo["login"].(string)
    if !ok {
        http.Error(w, "Username not found", http.StatusInternalServerError)
        return
    }

    email, ok := userInfo["email"].(string)
    if !ok || email == "" {
        // Optionally fetch the email directly from GitHub API if it's not public
		fmt.Println("Email is private, fetching from GitHub API...")
        email, err = fetchGitHubEmail(accessToken)
        if err != nil {
            http.Error(w, "Failed to retrieve email", http.StatusInternalServerError)
            return
        } 
    }

	exists, err := database.CheckEmailExists(email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		err = database.InsertUser(email, username, "")
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}
	}

	/// Create a unique session token using gofrs/uuid
	sessionToken, err := uuid.NewV4() // This generates a new UUID version 4
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Convert UUID to string for storage
	err = database.StoreSessionToken(username, sessionToken.String())
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Expires:  time.Now().Add(1 * time.Hour), // 1 hour only
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fetchGitHubEmail(accessToken string) (string, error) {
    req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var emails []struct {
        Email   string `json:"email"`
        Primary bool   `json:"primary"`
        Verified bool  `json:"verified"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
        return "", err
    }

    // Check for the first verified, primary email
    for _, email := range emails {
        if email.Primary && email.Verified {
            return email.Email, nil
        }
    }
    return "", fmt.Errorf("no primary verified email found")
}