package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"url-shortener/db/entities"
	"url-shortener/dto"
	"url-shortener/initializers"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShortenHandler(c *gin.Context) {
	var createShortenRequest dto.CreateShortenRequest
	err := c.ShouldBindJSON(&createShortenRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "url is required",
		})
		return
	}

	maxAttempts := 10
	var generatedCode string

	for i := 0; i < maxAttempts; i++ {
		generatedCode = services.GenerateShortenURL(6)

		var existing entities.ShortenedUrl
		err := initializers.DB.Where("short_url = ?", generatedCode).First(&existing).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			break // valid to save
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong wile generating a shortened url, try again later.",
			})
			return

		}

		if i == maxAttempts-1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("Failed to generate a shortened url within %d tries", maxAttempts),
			})
			return

		}
	}

	url := entities.ShortenedUrl{
		ShortUrl:       generatedCode,
		FullUrl:        createShortenRequest.Url,
		MaxUsageCount:  createShortenRequest.MaxUsageCount,
		ExpirationTime: createShortenRequest.Expiration,
	}

	initializers.DB.Create(&url)

	c.JSON(http.StatusCreated, gin.H{
		"shortUrl": url.ShortUrl,
	})
}

func RedirectHandler(c *gin.Context) {
	codeParam := c.Param("code")

	url := entities.ShortenedUrl{}
	result := initializers.DB.Where("short_url = ?", codeParam).First(&url)
	if result.Error != nil {
		// We don't want to give the specific error here to not give
		// insight in what went wrong to an end user.

		c.JSON(http.StatusNotFound, gin.H{
			"message": "shortened url not found",
		})
		return
	}

	// Check if URL has expired
	if time.Now().After(url.ExpirationTime) {
		c.JSON(http.StatusGone, gin.H{
			"message": "shortened url has expired",
		})
		return
	}

	// Check if usage limit has been reached
	if url.UsageCount >= url.MaxUsageCount {
		c.JSON(http.StatusGone, gin.H{
			"message": "shortened url usage limit exceeded",
		})
		return
	}

	// Increment usage count
	initializers.DB.Model(&url).Update("usage_count", url.UsageCount+1)

	c.Redirect(http.StatusFound, url.FullUrl)
}

func HealthStatus(c *gin.Context) {
	// TODO: check db status

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
