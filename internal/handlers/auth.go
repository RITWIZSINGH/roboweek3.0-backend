package handlers

import (
	"encoding/json"
	"net/http"

	"firebase.google.com/go/v4/auth"
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

	//Create custom token
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

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request)  {
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//Get user by email
	user,err := h.authClient.GetUserByEmail(r.Context(),req.Email)
	if err != nil {
		http.Error(w, "Invalid Credentials",http.StatusUnauthorized)
		return
	}

	//create custom token
	token, err := h.authClient.CustomToken(r.Context(), user.UID)
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	response := &models.AuthResponse{
		Token: token,
		User: &models.User{
			ID: user.UID,
			Email: user.Email,
			Name: user.DisplayName,
		},
	} 

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	
}

func (h *AuthHandler) GoogleSignIn(w http.ResponseWriter,r *http.Request)  {
	url := h.config.GoogleOAuthConfig.AuthCodeURL("state")
	http.Redirect(w,r,url,http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request)  {
	code := r.URL.Query().Get("code")

	token,err := h.config.GoogleOAuthConfig.Exchange(r.Context(),code)
	if err != nil {
		http.Error(w,"Failed to exchange token", http.StatusInternalServerError)
		return
	}

	//Verify ID token with Firebase
	idToken := token.Extra("id_token").(string)
	firebaseToken, err := h.authClient.VerifyIDToken(r.Context(),idToken)
	if err != nil {
		http.Error(w, "Invalid ID token", http.StatusUnauthorized)
		return
	} 

	//Get user details
	user,err := h.authClient.GetUser(r.Context(),firebaseToken.UID)
	if err != nil {
		http.Error(w, err.Error(),http.StatusInternalServerError)
		return
	}

	response := &models.AuthResponse{
		Token: idToken,
		User: &models.User{
			ID: user.UID,
			Email: user.Email,
			Name: user.DisplayName,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
