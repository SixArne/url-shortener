package services

import (
	"log"
	"time"
	"url-shortener/db/entities"
	"url-shortener/initializers"
)

func CleanupDatabase() {
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	now := time.Now()

	result := initializers.DB.Where(
		"created_at < ? OR usage_count >= max_usage_count OR expiration_time < ?",
		oneMonthAgo, now,
	).Delete(&entities.ShortenedUrl{})

	if result.Error != nil {
		log.Fatal(result.Error)
	}
}
