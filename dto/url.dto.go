package dto

import "time"

type CreateShortenRequest struct {
	Url           string    `json:"url" binding:"required"`
	Expiration    time.Time `json:"expiration"`
	MaxUsageCount uint      `json:"maxUsageCount" default:"0"`
}
