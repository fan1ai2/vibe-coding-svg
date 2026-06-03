package ai

import "context"

// IconCandidate is a single AI-generated SVG icon.
type IconCandidate struct {
	Name       string   `json:"name"`
	SvgContent string   `json:"svg_content"`
	Tags       []string `json:"tags"`
}

// Provider defines the AI icon generation interface.
type Provider interface {
	Generate(ctx context.Context, prompt, style string) ([]IconCandidate, error)
}
