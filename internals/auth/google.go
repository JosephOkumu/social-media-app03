package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// generateStateToken creates a random state token for OAuth flow
func generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	stateTokens[state] = time.Now().Add(15 * time.Minute)
	return state, nil
}

// InitiateGoogleAuth starts the Google OAuth flow
func InitiateGoogleAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateStateToken()
	if err != nil {
		log.Printf("Error generating state token: %v", err)
		http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
		return
	}

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?"+
			"client_id=%s"+
			"&redirect_uri=%s"+
			"&response_type=code"+
			"&scope=email+profile"+
			"&state=%s"+
			"&access_type=offline"+
			"&prompt=consent",
		url.QueryEscape(googleConfig.ClientID),
		url.QueryEscape(googleConfig.RedirectURI),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes the callback from Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Extract state and code from query parameters
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Validate state token
	expiry, exists := stateTokens[state]
	if !exists || time.Now().After(expiry) {
		log.Printf("Invalid state token or token expired")
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}
	delete(stateTokens, state)

	// Exchange code for token
	tokenResponse, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info using the access token
	userInfo, err := getGoogleUserInfo(tokenResponse.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Check if user exists in our database
	users := ReadfromDb()
	var existingUser *User
	for _, user := range users {
		if user.Email == userInfo.Email {
			existingUser = &user
			break
		}
	}

	var user User
	if existingUser == nil {
		// Create new user
		user = User{
			Email:     userInfo.Email,
			UserName:  generateUsername(userInfo),
			Password:  "", // Google-authenticated users don't need a password
			CreatedAt: time.Now(),
		}
		err = SaveUserToDb(user)
		if err != nil {
			log.Printf("Error saving user to database: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else {
		user = *existingUser
	}

	// Create session
	session := store.CreateSession(user.ID, user.UserName, r.RemoteAddr)
	if session == nil {
		log.Printf("Error creating session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// exchangeCodeForToken exchanges the authorization code for tokens
func exchangeCodeForToken(code string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", googleConfig.ClientID)
	data.Set("client_secret", googleConfig.ClientSecret)
	data.Set("redirect_uri", googleConfig.RedirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResponse GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

// getGoogleUserInfo retrieves the user's information from Google
func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateUsername generates a username based on the user's information
func generateUsername(userInfo *GoogleUserInfo) string {
	var base string

	// Prefer name components if available
	if userInfo.GivenName != "" && userInfo.FamilyName != "" {
		base = userInfo.GivenName + "." + userInfo.FamilyName
	} else if userInfo.Name != "" {
		base = userInfo.Name
	} else {
		base = strings.Split(userInfo.Email, "@")[0]
	}

	// Clean invalid characters using regex
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	clean := reg.ReplaceAllString(base, "-")

	// Normalize to lowercase
	clean = strings.ToLower(clean)

	// Trim length (3-20 characters is common for usernames)
	if len(clean) > 20 {
		clean = clean[:20]
	}

	// Ensure minimum length
	if len(clean) < 3 {
		clean = "user" + clean // Fallback prefix
		if len(clean) > 20 {
			clean = clean[:20]
		}
	}

	return clean
}
