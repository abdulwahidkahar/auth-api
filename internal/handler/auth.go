package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	var emailExists string
	err = ah.db.QueryRow("SELECT email FROM users WHERE email = $1", email).Scan(&emailExists)
	if err == nil {
		writeError(w, http.StatusConflict, "Email already registered")
		return
	}

	_, err = ah.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, string(passwordHash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error saving user to database")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully", "email": email})
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	var passwordHash string
	err := ah.db.QueryRow("SELECT password FROM users WHERE email = $1", email).Scan(&passwordHash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := generateToken(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Login successful", "email": email, "token": token})

}

func (ah *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	tokenStr := parts[1]
	secret := os.Getenv("JWT_SECRET")

	token, _ := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	claims, _ := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)
	userId := `SELECT id FROM users WHERE email = $1 `
	var id int
	err := ah.db.QueryRow(userId, email).Scan(&id)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "User not found")
		return
	}
	res := UserResponse{
		ID:    id,
		Email: email,
	}
	writeJSON(w, http.StatusOK, res)
}

func (ah *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Server is running"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func generateToken(email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}
