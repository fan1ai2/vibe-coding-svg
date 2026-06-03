package model

import "time"

type Icon struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	SvgContent    string    `json:"svg_content,omitempty"`
	IsPublic      bool      `json:"is_public"`
	DownloadCount int       `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Tags          []Tag     `json:"tags,omitempty"`
	Colors        []string  `json:"colors,omitempty"`
	Theme         string    `json:"theme,omitempty"`
}

type Tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Type       string `json:"type"`
	UsageCount int    `json:"usage_count"`
}
