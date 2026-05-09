package main

import (
	"auth-api/internal/database"
	"auth-api/internal/handler"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	r := gin.Default()

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

	r.GET("/gin", func(c *gin.Context) {
		c.JSON(http.StatusOK, "HELLO INI PAKE GIN")
	})

	authHandler := handler.NewAuthHandler(db)

	r.GET("/health", authHandler.Health)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(handler.JWTMiddleware)
	{
		api.GET("/profile", authHandler.Profile)
	}

	r.Run()
}
