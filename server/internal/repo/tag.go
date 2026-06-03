package repo

import (
	"database/sql"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type TagRepo struct{ db *sql.DB }

func NewTagRepo(db *sql.DB) *TagRepo { return &TagRepo{db} }

func (r *TagRepo) FindOrCreate(name, tagType string) (*model.Tag, error) {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")

	tag := &model.Tag{}
	err := r.db.QueryRow(
		`INSERT INTO tags (name, slug, type) VALUES ($1,$2,$3)
		 ON CONFLICT (slug) DO UPDATE SET usage_count = tags.usage_count + 1
		 RETURNING id, name, slug, type, usage_count`,
		name, slug, tagType,
	).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Type, &tag.UsageCount)
	return tag, err
}

func (r *TagRepo) FindBySlug(slug string) (*model.Tag, error) {
	tag := &model.Tag{}
	err := r.db.QueryRow(
		`SELECT id, name, slug, type, usage_count FROM tags WHERE slug=$1`, slug,
	).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Type, &tag.UsageCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tag, err
}

func (r *TagRepo) List(sortBy string, limit int) ([]model.Tag, error) {
	order := "usage_count DESC"
	if sortBy == "name" {
		order = "name ASC"
	}
	rows, err := r.db.Query(
		`SELECT id, name, slug, type, usage_count FROM tags ORDER BY `+order+` LIMIT $1`, limit,
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

func (r *TagRepo) IncrementUsage(tagID string) error {
	_, err := r.db.Exec(`UPDATE tags SET usage_count = usage_count + 1 WHERE id=$1`, tagID)
	return err
}

func (r *TagRepo) FindByIDs(ids []string) ([]model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(
		`SELECT id, name, slug, type, usage_count FROM tags WHERE id = ANY($1)`, pqArray(ids),
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

func pqArray(ids []string) interface{} {
	return "{" + strings.Join(ids, ",") + "}"
}
