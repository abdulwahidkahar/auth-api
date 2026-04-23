package handler

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (ah *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)

	email := body["email"]
	password := body["password"]

	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "Email and Password are required")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully", "email": email})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
