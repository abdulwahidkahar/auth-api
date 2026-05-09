package handler

import (
	"auth-api/model"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (ah *AuthHandler) Register(c *gin.Context) {

	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and Password are required"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	var emailExists string
	err = ah.db.QueryRow("SELECT email FROM users WHERE email = $1", req.Email).Scan(&emailExists)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	_, err = ah.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", req.Email, string(passwordHash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user to database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "email": req.Email})
}

func (ah *AuthHandler) Login(c *gin.Context) {

	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email and password are required",
		})
		return
	}

	var (
		id           int
		passwordHash string
	)

	query := `
		SELECT id, password
		FROM users
		WHERE email = $1
	`

	err := ah.db.QueryRow(query, req.Email).
		Scan(&id, &passwordHash)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	token, err := generateToken(id, req.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generating token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

func (ah *AuthHandler) Profile(c *gin.Context) {

	userID := int(c.MustGet("id").(float64))

	query := `SELECT id, email FROM users WHERE id = $1`

	var user model.UserResponse

	err := ah.db.QueryRow(query, int(userID)).
		Scan(&user.ID, &user.Email)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ah *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Server is running"})
}

func writeJSON(c *gin.Context, status int, data any) {
	c.Header("Content-Type", "application/json")
	c.Status(status)
	json.NewEncoder(c.Writer).Encode(data)
}

func writeError(c *gin.Context, status int, message string) {
	writeJSON(c, status, map[string]string{"error": message})
}

func generateToken(id int, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}
