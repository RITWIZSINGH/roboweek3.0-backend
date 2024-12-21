package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"roboweek3.0-backend/internal/config"
	"roboweek3.0-backend/internal/models"
)

type AuthHandler struct {
	config     *config.Config
	authClient *auth.Client
}

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(cfg *config.Config, authClient *auth.Client) *AuthHandler {
	return &AuthHandler{
		config:     cfg,
		authClient: authClient,
	}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(req.Email).Password(req.Password).DisplayName(req.Name)

	userRecord, err := h.authClient.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.authClient.CustomToken(r.Context(), userRecord.UID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.AuthResponse{
		Token: token,
		User: &models.User{
			ID:    userRecord.UID,
			Email: userRecord.Email,
			Name:  userRecord.DisplayName,
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authClient.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.authClient.CustomToken(r.Context(), user.UID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.AuthResponse{
		Token: token,
		User: &models.User{
			ID:    user.UID,
			Email: user.Email,
			Name:  user.DisplayName,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) GoogleSignIn(w http.ResponseWriter, r *http.Request) {
	url := h.config.GoogleOAuthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")

    token, err := h.config.GoogleOAuthConfig.Exchange(r.Context(), code)
    if err != nil {
        log.Printf("Failed to exchange token: %v\n", err)
        http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
        return
    }

    // Get user info from Google
    oauth2Service, err := oauth2.NewService(r.Context(), 
        option.WithTokenSource(h.config.GoogleOAuthConfig.TokenSource(r.Context(), token)))
    if err != nil {
        log.Printf("Failed to create OAuth2 service: %v\n", err)
        http.Error(w, "Failed to get user info", http.StatusInternalServerError)
        return
    }

    userInfo, err := oauth2Service.Userinfo.Get().Do()
    if err != nil {
        log.Printf("Failed to get user info: %v\n", err)
        http.Error(w, "Failed to get user info", http.StatusInternalServerError)
        return
    }

    // Check if user exists in Firebase
    user, err := h.authClient.GetUserByEmail(r.Context(), userInfo.Email)
    if err != nil {
        // User doesn't exist, create new user in Firebase
        params := (&auth.UserToCreate{}).
            Email(userInfo.Email).
            DisplayName(userInfo.Name).
            PhotoURL(userInfo.Picture)

        user, err = h.authClient.CreateUser(r.Context(), params)
        if err != nil {
            log.Printf("Failed to create user: %v\n", err)
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }
    }

    // Create custom token
    customToken, err := h.authClient.CustomToken(r.Context(), user.UID)
    if err != nil {
        log.Printf("Error creating custom token: %v\n", err)
        http.Error(w, "Error creating custom token", http.StatusInternalServerError)
        return
    }

    // Create response
    response := &models.AuthResponse{
        Token: customToken,
        User: &models.User{
            ID:    user.UID,
            Email: user.Email,
            Name:  user.DisplayName,
        },
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
