package initializers

import (
	"log"
	"url-shortener/db/entities"
)

func InitMigration() {
	err := DB.AutoMigrate(entities.ShortenedUrl{})
	if err != nil {
		log.Fatal("Failed to migrate")
	}
}
