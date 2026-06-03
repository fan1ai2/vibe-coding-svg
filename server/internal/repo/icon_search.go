package repo

import (
	"fmt"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type IconSearchParams struct {
	Query  string
	Tags   []string
	Color  string
	Theme  string
	Sort   string
	Limit  int
	Offset int
	UserID string // If set, include user's private icons in results
}

func (r *IconRepo) Search(params IconSearchParams) ([]*model.Icon, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	// Visibility
	if params.UserID != "" {
		where = append(where, fmt.Sprintf("(i.is_public=true OR i.user_id=$%d)", argIdx))
		args = append(args, params.UserID)
		argIdx++
	} else {
		where = append(where, "i.is_public=true")
	}

	// Keyword search
	if params.Query != "" {
		where = append(where, fmt.Sprintf(
			`(to_tsvector('simple', regexp_replace(i.name, '[-.]', ' ', 'g')) @@ plainto_tsquery('simple', regexp_replace($%d, '[-.]', ' ', 'g')))`, argIdx))
		args = append(args, params.Query)
		argIdx++
	}

	// Tag filter (AND)
	if len(params.Tags) > 0 {
		placeholders := make([]string, len(params.Tags))
		for i, tag := range params.Tags {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, tag)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			`i.id IN (SELECT it.icon_id FROM icon_tags it
			           JOIN tags t ON t.id = it.tag_id
			           WHERE t.slug IN (%s)
			           GROUP BY it.icon_id HAVING COUNT(DISTINCT t.slug) = %d)`,
			strings.Join(placeholders, ","), len(params.Tags)))
	}

	// Color filter (HSL hue ±15)
	if params.Color != "" {
		hue := colorToHueExpr(params.Color)
		where = append(where, fmt.Sprintf(
			`i.id IN (SELECT ic.icon_id FROM icon_colors ic WHERE %s)`, hue))
	}

	// Theme filter
	if params.Theme != "" {
		where = append(where, fmt.Sprintf("i.id IN (SELECT it.icon_id FROM icon_themes it WHERE it.theme_name=$%d)", argIdx))
		args = append(args, params.Theme)
		argIdx++
	}

	// Sort
	order := "i.created_at DESC"
	switch params.Sort {
	case "popular":
		order = "i.download_count DESC"
	case "newest":
		order = "i.created_at DESC"
	}

	// Pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}

	query := fmt.Sprintf(
		`SELECT i.id, i.user_id, i.name, i.svg_content, i.is_public, i.download_count, i.created_at, i.updated_at
		 FROM icons i WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), order, argIdx, argIdx+1,
	)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.Icon, 0)
	for rows.Next() {
		icon := &model.Icon{}
		if err := rows.Scan(&icon.ID, &icon.UserID, &icon.Name, &icon.SvgContent,
			&icon.IsPublic, &icon.DownloadCount, &icon.CreatedAt, &icon.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, icon)
	}
	return list, rows.Err()
}

// colorToHueExpr builds a SQL expression matching HSL hue ±15 degrees.
// Accepts hex (#FF0000) or named colors mapped to approximate hue.
func colorToHueExpr(color string) string {
	h := hexToHue(color)
	lo := h - 15
	hi := h + 15
	if lo < 0 {
		return fmt.Sprintf(
			`(hue_from_hex(ic.color_hex) BETWEEN 0 AND %d OR hue_from_hex(ic.color_hex) BETWEEN %d AND 360)`,
			hi, lo+360)
	}
	if hi > 360 {
		return fmt.Sprintf(
			`(hue_from_hex(ic.color_hex) BETWEEN %d AND 360 OR hue_from_hex(ic.color_hex) BETWEEN 0 AND %d)`,
			lo, hi-360)
	}
	return fmt.Sprintf(`(hue_from_hex(ic.color_hex) BETWEEN %d AND %d)`, lo, hi)
}

func hexToHue(hex string) int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 0
	}
	r := hexByte(hex[0:2])
	g := hexByte(hex[2:4])
	b := hexByte(hex[4:6])
	max := max(r, max(g, b))
	min := min(r, min(g, b))
	if max == min {
		return 0
	}
	delta := max - min
	var h float64
	switch max {
	case r:
		h = float64(g-b)/float64(delta) + 0
	case g:
		h = float64(b-r)/float64(delta) + 2
	case b:
		h = float64(r-g)/float64(delta) + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return int(h)
}

func hexByte(s string) int {
	if len(s) < 2 {
		return 0
	}
	var v int
	fmt.Sscanf(s, "%02x", &v)
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

