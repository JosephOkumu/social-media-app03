package auth

import "time"

type User struct {
	ID        int
	UserName  string
	Email     string
	Password  string
	CreatedAt time.Time
}

var store = NewSessionStore()

type contextKey string

const UserSessionKey contextKey = "userSession"

// PageData represents the data structure we'll pass to our templates
type PageData struct {
	IsLoggedIn bool
	UserName   string
}

// GoogleConfig holds the configuration for Google OAuth
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// GoogleUserInfo represents the user information received from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GoogleTokenResponse represents the OAuth token response
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// Create a map to store state tokens to prevent CSRF attacks
var stateTokens = make(map[string]time.Time)

// Initialize Google configuration
var googleConfig = &GoogleConfig{
	ClientID:     "163539294903-216klc89htmsk9kpigk6apf5q15n5e1b.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURI:  "http://localhost:8080/auth/google/callback",
}
