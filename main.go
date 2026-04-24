package main

import (
	"auth-api/internal/database"
	"auth-api/internal/handler"
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("error loading env")
	}

	db, err := database.NewPostgresDB()
	if err != nil {
		fmt.Println("error connecting to database:", err)
	}

	if err == nil {
		fmt.Println("Connected to PostgreSQL database successfully!")
	}
	defer db.Close()

	authHandler := handler.NewAuthHandler(db)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.ListenAndServe(":8080", nil)
}
