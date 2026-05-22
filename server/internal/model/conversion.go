package model

import "time"

type Conversion struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	Status       string     `json:"status" db:"status"`
	OriginalURL  string     `json:"original_url,omitempty" db:"original_url"`
	SVGURL       *string    `json:"svg_url,omitempty" db:"svg_url"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	FileSizeIn   int64      `json:"file_size_in" db:"file_size_in"`
	FileSizeOut  *int64     `json:"file_size_out" db:"file_size_out"`
	PathCount    *int       `json:"path_count" db:"path_count"`
	ColorCount   *int       `json:"color_count" db:"color_count"`
	FormatIn     string     `json:"format_in" db:"format_in"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)
