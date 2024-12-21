package config

import (
    "log"
    "os"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "github.com/joho/godotenv"
)

type Config struct {
    GoogleOAuthConfig *oauth2.Config
}

func NewConfig() *Config {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not found")
    }

    return &Config{
        GoogleOAuthConfig: &oauth2.Config{
            ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
            ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
            RedirectURL:  "http://localhost:8000/auth/google/callback",
            Scopes: []string{
                "https://www.googleapis.com/auth/userinfo.email",
                "https://www.googleapis.com/auth/userinfo.profile",
            },
            Endpoint: google.Endpoint,
        },
    }
}