package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type User struct {
	Name      string `json:"Name"`
	Post      string `json:"Post"`
	ImageURL  string `json:"ProfilePicture"`
	Github    string `json:"Github"`
	LinkedIn  string `json:"LinkedIn"`
	Instagram string `json:"Instagram"`
	TechStack string `json:"TechStack"`
}

var data []User

func init() {
	// Load JSON data from a file
	fileContent, err := ioutil.ReadFile("test.json")
	if err != nil {
		log.Fatalf("Error reading test.json: %v", err)
	}

	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		log.Fatalf("Error parsing test.json: %v", err)
	}
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) // Send all users' data as JSON
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome! Use /users to get all data.
        Use /auth/signup to sign up.
        Use /auth/signin to sign in.
        Use /auth/google/signin to sign in with Google.
        Use /auth/google/callback to get the Google sign in callback.`))
}

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
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false, // Must be false when AllowedOrigins is "*"
		MaxAge:           300,   // Maximum value not ignored by any of major browsers
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
	r.Get("/users", getAllUsers)
	r.Get("/", home)

	port := 8000
	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
