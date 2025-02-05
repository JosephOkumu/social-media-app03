package auth

import (
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

// InitiateFacebookAuth starts the Facebook OAuth flow
func InitiateFacebookAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateStateToken()
	if err != nil {
		http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
		return
	}

	authURL := fmt.Sprintf(
		"https://www.facebook.com/v12.0/dialog/oauth?"+
			"client_id=%s"+
			"&redirect_uri=%s"+
			"&response_type=code"+
			"&scope=%s"+
			"&state=%s",
		url.QueryEscape(facebookConfig.ClientID),
		url.QueryEscape(facebookConfig.RedirectURI),
		url.QueryEscape(strings.Join(facebookConfig.Scopes, " ")),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleFacebookCallback processes the callback from Facebook
func HandleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Facebook callback...")

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
	tokenResponse, err := exchangeCodeForFacebookToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info using the access token
	userInfo, err := getFacebookUserInfo(tokenResponse.AccessToken)
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
			UserName:  generateUsernameFromFacebook(userInfo),
			Password:  "", // Facebook-authenticated users don't need a password
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

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// exchangeCodeForFacebookToken exchanges the authorization code for tokens
func exchangeCodeForFacebookToken(code string) (*FacebookTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", facebookConfig.ClientID)
	data.Set("client_secret", facebookConfig.ClientSecret)
	data.Set("redirect_uri", facebookConfig.RedirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://graph.facebook.com/v12.0/oauth/access_token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var tokenResponse FacebookTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Printf("Error unmarshalling token response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal token response: %v", err)
	}

	return &tokenResponse, nil
}

// getFacebookUserInfo retrieves the user's information from Facebook
func getFacebookUserInfo(accessToken string) (*FacebookUserInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v12.0/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Specify the fields you want to retrieve
	q := req.URL.Query()
	q.Add("fields", "id,email,name")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var userInfo FacebookUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		log.Printf("Error unmarshalling user info: %v", err)
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	return &userInfo, nil
}

// generateUsernameFromFacebook generates a username based on the user's information
func generateUsernameFromFacebook(userInfo *FacebookUserInfo) string {
	var base string

	// Use the user's name or email as the base for the username
	if userInfo.Name != "" {
		base = userInfo.Name
	} else {
		base = strings.Split(userInfo.Email, "@")[0]
	}

	// Clean invalid characters using regex
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	clean := reg.ReplaceAllString(base, "-")

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
