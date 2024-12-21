package models

type User struct {
	ID string `json:"id"`
	Email string `json:"email"`
	Name string `json:"name"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User *User `json:"user"`
}