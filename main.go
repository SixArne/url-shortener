package main

import (
	"log"
	"url-shortener/handlers"
	"url-shortener/initializers"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func init() {
	//initializers.InitEnvironment()
	initializers.InitDatabase()
	initializers.InitMigration()
}

func main() {
	router := gin.Default()
	c := cron.New()

	_, err := c.AddFunc("@daily", services.CleanupDatabase)
	if err != nil {
		log.Fatal("Failed to add daily cron job")
	}

	router.GET("/status", handlers.HealthStatus)
	router.POST("/shorten", handlers.ShortenHandler)
	router.GET("/:code", handlers.RedirectHandler)

	err = router.Run(":8080")
	if err != nil {
		log.Fatal("Error starting server", err)
	}
}
