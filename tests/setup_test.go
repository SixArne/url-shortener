package tests

import (
	"os"
	"testing"
	"url-shortener/db/entities"
	"url-shortener/handlers"
	"url-shortener/initializers"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Setup: runs once before all tests
	setupTestDB()

	// Run all tests
	code := m.Run()

	// Teardown: runs once after all tests
	// (optional cleanup if needed)

	os.Exit(code)
}

func setupTestDB() {
	var err error
	initializers.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}

	// Run migrations
	err = initializers.DB.AutoMigrate(&entities.ShortenedUrl{})
	if err != nil {
		panic("failed to migrate test database")
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/shorten", handlers.ShortenHandler)
	router.GET("/:code", handlers.RedirectHandler)

	return router
}

func cleanupTestDB() {
	initializers.DB.Exec("DELETE FROM shortened_urls")
}
