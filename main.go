package main

import (
	"auth-api/internal/handler"
	"net/http"
)

func main() {
	authHandler := handler.NewAuthHandler()
	http.HandleFunc("/register", authHandler.RegisterHandler)
	http.ListenAndServe(":8080", nil)
}
