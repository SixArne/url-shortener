package entities

import (
	"time"

	"gorm.io/gorm"
)

type ShortenedUrl struct {
	gorm.Model

	ShortUrl       string `gorm:"primary_key"`
	FullUrl        string
	UsageCount     uint
	MaxUsageCount  uint
	ExpirationTime time.Time
}
