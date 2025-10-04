package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shortener/db/entities"
	"url-shortener/dto"
	"url-shortener/initializers"

	"github.com/stretchr/testify/assert"
)

func TestCreateShortUrl(t *testing.T) {
	defer cleanupTestDB()

	router := setupRouter()

	// Create request body
	requestBody := dto.CreateShortenRequest{
		Url:           "https://example.com/very/long/url/to/shorten",
		Expiration:    time.Now().Add(24 * time.Hour),
		MaxUsageCount: 10,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response status
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "shortUrl")

	shortUrl := response["shortUrl"]
	assert.NotEmpty(t, shortUrl)

	// Validate in database
	var savedUrl entities.ShortenedUrl
	result := initializers.DB.Where("short_url = ?", shortUrl).First(&savedUrl)
	assert.NoError(t, result.Error)
	assert.Equal(t, requestBody.Url, savedUrl.FullUrl)
	assert.Equal(t, requestBody.MaxUsageCount, savedUrl.MaxUsageCount)
	assert.Equal(t, uint(0), savedUrl.UsageCount)
	assert.WithinDuration(t, requestBody.Expiration, savedUrl.ExpirationTime, time.Second)
}

func TestRedirectShortUrl(t *testing.T) {
	defer cleanupTestDB()

	router := setupRouter()

	// Create a shortened URL in the database
	shortCode := "test123"
	originalUrl := "https://example.com/redirect-target"
	url := entities.ShortenedUrl{
		ShortUrl:       shortCode,
		FullUrl:        originalUrl,
		UsageCount:     0,
		MaxUsageCount:  5,
		ExpirationTime: time.Now().Add(24 * time.Hour),
	}

	initializers.DB.Create(&url)

	// Make GET request to redirect endpoint
	req, _ := http.NewRequest("GET", "/"+shortCode, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert redirect response
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, originalUrl, w.Header().Get("Location"))

	// Verify usage count incremented in database
	var updatedUrl entities.ShortenedUrl
	initializers.DB.Where("short_url = ?", shortCode).First(&updatedUrl)
	assert.Equal(t, uint(1), updatedUrl.UsageCount)
}

func TestRedirectExpiredUrl(t *testing.T) {
	defer cleanupTestDB()

	router := setupRouter()

	// Create an expired shortened URL in the database
	shortCode := "expired1"
	url := entities.ShortenedUrl{
		ShortUrl:       shortCode,
		FullUrl:        "https://example.com/expired",
		UsageCount:     0,
		MaxUsageCount:  5,
		ExpirationTime: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	initializers.DB.Create(&url)

	// Make GET request to redirect endpoint
	req, _ := http.NewRequest("GET", "/"+shortCode, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert HTTP 410 Gone response
	assert.Equal(t, http.StatusGone, w.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "shortened url has expired", response["message"])
}

func TestRedirectMaxUsageExceeded(t *testing.T) {
	defer cleanupTestDB()

	router := setupRouter()

	// Create a shortened URL that has reached its max usage count
	shortCode := "maxed123"
	url := entities.ShortenedUrl{
		ShortUrl:       shortCode,
		FullUrl:        "https://example.com/maxed-out",
		UsageCount:     3,
		MaxUsageCount:  3, // Already at limit
		ExpirationTime: time.Now().Add(24 * time.Hour),
	}

	initializers.DB.Create(&url)

	// Make GET request to redirect endpoint
	req, _ := http.NewRequest("GET", "/"+shortCode, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert HTTP 410 Gone response
	assert.Equal(t, http.StatusGone, w.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "shortened url usage limit exceeded", response["message"])
}
