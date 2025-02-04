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

// GitHubConfig holds the configuration for GitHub OAuth
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// FacebookConfig holds the configuration for Facebook OAuth
type FacebookConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string 
}

// GoogleUserInfo represents the user information received from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type GitHubUserInfo struct {
	Username      string `json:"login"`          // GitHub username (login)
	Email         string `json:"email"`          // Public email, if accessible
	VerifiedEmail bool   `json:"verified_email"` // Whether the email is verified
	Name          string `json:"name"`           // Full name (may be empty)
	Location      string `json:"location"`       // Location (if available)
}

type FacebookUserInfo struct {
	ID      string `json:"id"`       // Unique Facebook ID
	Email   string `json:"email"`    
	Name    string `json:"name"`     
}

// GoogleTokenResponse represents the OAuth token response
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// GitHUbTokenResponse represents the OAuth token response
type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json: "scope"`
	TokenType   string `json:"token_type"`
}

// FacebookTokenResponse represents the OAuth token response from Facebook
type FacebookTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // Token expiration time in seconds
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// Create a map to store state tokens to prevent CSRF attacks
var stateTokens = make(map[string]time.Time)

// Initialize Google configuration
var googleConfig = &GoogleConfig{
	ClientID:     "163539294903-216klc89htmsk9kpigk6apf5q15n5e1b.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURI:  "http://localhost:8080/auth/google/callback",
}

// Initialize Google configuration
var githubConfig = &GitHubConfig{
	ClientID:     "Ov23liYc9Bf6D3ehd3Yj",
	ClientSecret: "befd89c13ee39ffb0326fe3e28ec63d41fe73d3b",
	RedirectURI:  "http://localhost:8080/auth/github/callback",
}

