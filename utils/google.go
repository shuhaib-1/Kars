package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv" // Make sure to import this package for loading .env files
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Initialize the environment variables
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// Google OAuth configuration function
func GetGoogleOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatalf("Missing GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET in environment variables")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "/kars.in.net/api/user/google_signup/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// Get Google Auth URL
func GetGoogleAuthURL(state string) string {
	config := GetGoogleOAuthConfig() // Fetch configuration dynamically
	fmt.Println("Client ID:", config.ClientID)
	fmt.Println("Client Secret:", config.ClientSecret)
	return config.AuthCodeURL(state)
}

// Exchange the authorization code for a token
func ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	config := GetGoogleOAuthConfig() // Fetch configuration dynamically
	return config.Exchange(ctx, code)
}
