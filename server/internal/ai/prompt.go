package ai

import (
	"database/sql"
	"fmt"
	"strings"
)

// PromptBuilder constructs system prompts for AI icon generation.
type PromptBuilder struct {
	db *sql.DB
}

func NewPromptBuilder(db *sql.DB) *PromptBuilder {
	return &PromptBuilder{db: db}
}

// BuildSystemPrompt creates the 4-layer system prompt, injecting library
// context (top tags + colors) from the database.
func (b *PromptBuilder) BuildSystemPrompt(style string) string {
	tags := b.topTags(10)
	colors := b.topColors(5)

	var sb strings.Builder

	// Layer 1: Role
	sb.WriteString("You are a professional UI icon designer specialized in creating minimal, pixel-perfect SVG icons for modern web and mobile applications.\n\n")

	// Layer 2: Format constraints
	defaultColor := "#F59E0B"
	if style == "filled" {
		defaultColor = "#F59E0B"
	}
	sb.WriteString("Generate SVG icons with these EXACT specifications:\n")
	sb.WriteString("- <svg> must have xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\" and NO width/height attributes\n")
	if style == "line" {
		sb.WriteString(fmt.Sprintf("- Line style icons: fill=\"none\", stroke=\"%s\", stroke-width=\"2\", stroke-linecap=\"round\", stroke-linejoin=\"round\"\n", defaultColor))
	} else {
		sb.WriteString("- Filled style icons: use solid fill colors, no stroke elements\n")
	}
	sb.WriteString("- Output pure SVG only — no markdown, no code fences, no HTML wrapper\n")
	sb.WriteString("- No external references (no <use>, no url() to external resources)\n")
	sb.WriteString("- Keep paths simple and clean — good icons use as few elements as possible\n\n")

	// Layer 3: Library style injection
	if len(tags) > 0 {
		sb.WriteString(fmt.Sprintf("Reference these common tags for icon style consistency: %s\n", strings.Join(tags, ", ")))
	}
	if len(colors) > 0 {
		sb.WriteString(fmt.Sprintf("Preferred color palette: %s\n", strings.Join(colors, ", ")))
	}
	sb.WriteString("\n")

	// Layer 4: Output format
	sb.WriteString("Return exactly 4 icons as a raw JSON array (no markdown fences, no extra text):\n")
	sb.WriteString(`[{"name":"icon-name","svg_content":"<svg viewBox=\"0 0 24 24\" ...>...</svg>","tags":["tag1","tag2"]}]` + "\n")
	sb.WriteString(`The "tags" field should suggest 2-3 appropriate tags for this icon.`)

	return sb.String()
}

func (b *PromptBuilder) topTags(limit int) []string {
	if b.db == nil {
		return nil
	}
	rows, err := b.db.Query(`SELECT name FROM tags ORDER BY usage_count DESC LIMIT $1`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tags = append(tags, name)
		}
	}
	return tags
}

var defaultColors = []string{"#F59E0B", "#10B981", "#EF4444", "#F59E0B", "#6B7280"}

func (b *PromptBuilder) topColors(limit int) []string {
	if b.db == nil {
		return defaultColors
	}
	rows, err := b.db.Query(`SELECT color_hex, COUNT(*) AS cnt FROM icon_colors GROUP BY color_hex ORDER BY cnt DESC LIMIT $1`, limit)
	if err != nil {
		return defaultColors
	}
	defer rows.Close()
	var colors []string
	for rows.Next() {
		var c string
		var cnt int
		if err := rows.Scan(&c, &cnt); err == nil {
			colors = append(colors, c)
		}
	}
	if len(colors) == 0 {
		return defaultColors
	}
	return colors
}
