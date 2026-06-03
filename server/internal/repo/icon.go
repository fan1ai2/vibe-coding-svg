package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/lib/pq"
)

type IconRepo struct{ db *sql.DB }

func NewIconRepo(db *sql.DB) *IconRepo { return &IconRepo{db} }

func (r *IconRepo) Create(userID, name, svgContent string, isPublic bool) (*model.Icon, error) {
	icon := &model.Icon{}
	err := r.db.QueryRow(
		`INSERT INTO icons (user_id, name, svg_content, is_public)
		 VALUES ($1,$2,$3,$4)
		 RETURNING id, user_id, name, svg_content, is_public, download_count, created_at, updated_at`,
		userID, name, svgContent, isPublic,
	).Scan(&icon.ID, &icon.UserID, &icon.Name, &icon.SvgContent,
		&icon.IsPublic, &icon.DownloadCount, &icon.CreatedAt, &icon.UpdatedAt)
	return icon, err
}

type BatchIcon struct {
	Name       string   `json:"name"`
	SvgContent string   `json:"svg_content"`
	IsPublic   bool     `json:"is_public"`
	Tags       []string `json:"tags"`
	Theme      string   `json:"theme"`
}

func (r *IconRepo) BatchCreate(userID string, icons []BatchIcon) ([]*model.Icon, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result := make([]*model.Icon, 0, len(icons))
	for _, ic := range icons {
		var icon model.Icon
		err := tx.QueryRow(
			`INSERT INTO icons (user_id, name, svg_content, is_public)
			 VALUES ($1,$2,$3,$4)
			 RETURNING id, user_id, name, svg_content, is_public, download_count, created_at, updated_at`,
			userID, ic.Name, ic.SvgContent, ic.IsPublic,
		).Scan(&icon.ID, &icon.UserID, &icon.Name, &icon.SvgContent,
			&icon.IsPublic, &icon.DownloadCount, &icon.CreatedAt, &icon.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("batch create icon %q: %w", ic.Name, err)
		}
		result = append(result, &icon)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *IconRepo) FindByID(id string) (*model.Icon, error) {
	icon := &model.Icon{}
	err := r.db.QueryRow(
		`SELECT id, user_id, name, svg_content, is_public, download_count, created_at, updated_at
		 FROM icons WHERE id=$1`, id,
	).Scan(&icon.ID, &icon.UserID, &icon.Name, &icon.SvgContent,
		&icon.IsPublic, &icon.DownloadCount, &icon.CreatedAt, &icon.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return icon, err
}

func (r *IconRepo) FindByUserID(userID string, limit, offset int) ([]*model.Icon, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, svg_content, is_public, download_count, created_at, updated_at
		 FROM icons WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
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

func (r *IconRepo) ListPublic(limit, offset int) ([]*model.Icon, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, svg_content, is_public, download_count, created_at, updated_at
		 FROM icons WHERE is_public=true ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
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

func (r *IconRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM icons WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *IconRepo) SetVisibility(id string, isPublic bool) error {
	_, err := r.db.Exec(`UPDATE icons SET is_public=$1, updated_at=$2 WHERE id=$3`, isPublic, time.Now(), id)
	return err
}

func (r *IconRepo) AttachTags(iconID string, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}
	vals := make([]string, 0, len(tagIDs))
	args := make([]interface{}, 0, len(tagIDs)*2)
	for i, tid := range tagIDs {
		vals = append(vals, fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2))
		args = append(args, iconID, tid)
	}
	_, err := r.db.Exec(
		`INSERT INTO icon_tags (icon_id, tag_id) VALUES `+strings.Join(vals, ",")+` ON CONFLICT DO NOTHING`,
		args...,
	)
	return err
}

func (r *IconRepo) AttachColors(iconID string, colors []string) error {
	if len(colors) == 0 {
		return nil
	}
	vals := make([]string, 0, len(colors))
	args := make([]interface{}, 0, len(colors)*3)
	for i, c := range colors {
		vals = append(vals, fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, iconID, c, "fill")
	}
	_, err := r.db.Exec(
		`INSERT INTO icon_colors (icon_id, color_hex, role) VALUES `+strings.Join(vals, ",")+` ON CONFLICT DO NOTHING`,
		args...,
	)
	return err
}

func (r *IconRepo) AttachTheme(iconID, theme string) error {
	if theme == "" {
		return nil
	}
	_, err := r.db.Exec(
		`INSERT INTO icon_themes (icon_id, theme_name) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		iconID, theme,
	)
	return err
}

func (r *IconRepo) LoadTags(iconID string) ([]model.Tag, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.name, t.slug, t.type, t.usage_count
		 FROM tags t JOIN icon_tags it ON t.id = it.tag_id
		 WHERE it.icon_id = $1`, iconID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := make([]model.Tag, 0)
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Type, &t.UsageCount); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *IconRepo) LoadColors(iconID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT color_hex FROM icon_colors WHERE icon_id=$1 ORDER BY color_hex`, iconID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	colors := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}
	return colors, rows.Err()
}

func (r *IconRepo) FindByIDs(ids []string) ([]*model.Icon, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(
		`SELECT id, user_id, name, svg_content, is_public, download_count, created_at, updated_at
		 FROM icons WHERE id = ANY($1) AND is_public=true`, pq.Array(ids),
	)
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
