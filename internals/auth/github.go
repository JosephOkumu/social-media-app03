package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func InitiateGitHubAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateStateToken()
	if err != nil {
		log.Printf("Error generating state token: %v", err)
		http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
		return
	}

	// Redirect to GitHub's authorization page
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=%s",
		url.QueryEscape(githubConfig.ClientID),
		url.QueryEscape(githubConfig.RedirectURI),
		url.QueryEscape(state),
		url.QueryEscape("user:email"),
	)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Extract state and code from query parameters
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Check if the state token is valid
	expiry, ok := stateTokens[state]
	if !ok || time.Now().After(expiry) {
		log.Printf("Invalid state token: %s", state)
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	delete(stateTokens, state)

	// Exchange the code for an access token
	tokenResponse, err := exchangeCodeForTokenGithub(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}
	// Get user info using the access token
	userInfo, err := getGithubUserInfo(tokenResponse.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	emails, err := getGithubUserEmails(tokenResponse.AccessToken)
	if err != nil {
		log.Printf("Error getting user emails: %v", err)
		http.Error(w, "Failed to get user emails", http.StatusInternalServerError)
		return
	}

	userInfo.Email = emails[0].Email

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
			UserName:  userInfo.Username,
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
		if oldsession, ok := store.GetSessionByUserId(user.ID); ok {
			store.DeleteSession(oldsession.ID)
		}
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

func exchangeCodeForTokenGithub(code string) (*GitHubTokenResponse, error) {
	// Create a POST request to exchange the code for an access token
	data := url.Values{}
	data.Set("client_id", githubConfig.ClientID)
	data.Set("client_secret", githubConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", githubConfig.RedirectURI)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Println("Error creating request: ", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body: ", err)
		return nil, err
	}

	// Parse the response body to get the access token
	var tokenResponse GitHubTokenResponse
	if err = json.Unmarshal(body, &tokenResponse); err != nil {
		log.Println("Error unmarshalling response body: ", err)
		return nil, err
	}
	return &tokenResponse, nil
}

func getGithubUserInfo(token string) (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response body to get the user info
	var userInfo GitHubUserInfo
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func getGithubUserEmails(token string) ([]GitHubEmail, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var emails []GitHubEmail
	if err = json.Unmarshal(body, &emails); err != nil {
		return nil, err
	}

	// Find the primary email
	for _, email := range emails {
		if email.Primary {
			return []GitHubEmail{email}, nil
		}
	}

	return nil, nil
}
