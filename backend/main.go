package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"

	"live-polling-tool/backend/config"
	"live-polling-tool/backend/routes"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	config.ConnectDB()

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Live Polling Tool API is running",
		})
	})

	routes.PollRoutes(r)

	fmt.Println("Server running on http://localhost:8081")

	r.Run(":8081")
}
