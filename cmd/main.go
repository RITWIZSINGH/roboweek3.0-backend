package main

import (
    "context"
    "log"
    "net/http"

    firebase "firebase.google.com/go/v4"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "google.golang.org/api/option"
    "roboweek3.0-backend/internal/config"
    "roboweek3.0-backend/internal/handlers"
)

func main() {
    //Initialize Firebase
    opt := option.WithCredentialsFile("./roboweek3-firebase-adminsdk-z27vj-d49ee40fe2.json")
    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        log.Fatalf("Error initializing Firebase app: %v\n", err)
    }

    //Get Firebase Auth Client
    authClient, err := app.Auth(context.Background())
    if err != nil {
        log.Fatalf("Error getting Auth client: %v\n", err)
    }

    r := chi.NewRouter()

    // CORS middleware with all origins allowed
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"},                                // Allow all origins
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: false,  // Must be false when AllowedOrigins is "*"
        MaxAge:           300,    // Maximum value not ignored by any of major browsers
    }))

    //Other Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    //Initialize config
    cfg := config.NewConfig()

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(cfg, authClient)

    //Routes
    r.Post("/auth/signup", authHandler.SignUp)
    r.Post("/auth/signin", authHandler.SignIn)
    r.Get("/auth/google/signin", authHandler.GoogleSignIn)
    r.Get("/auth/google/callback", authHandler.GoogleCallback)

    log.Fatal(http.ListenAndServe(":8000", r))
}