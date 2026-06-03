package repo

import (
	"database/sql"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type SavedSvgRepo struct{ db *sql.DB }

func NewSavedSvgRepo(db *sql.DB) *SavedSvgRepo { return &SavedSvgRepo{db} }

func (r *SavedSvgRepo) Create(userID, name, svgContent string) (*model.SavedSvg, error) {
	s := &model.SavedSvg{}
	err := r.db.QueryRow(
		`INSERT INTO saved_svgs (user_id, name, svg_content) VALUES ($1,$2,$3)
		 RETURNING id, user_id, name, svg_content, created_at, updated_at`,
		userID, name, svgContent,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.SvgContent, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *SavedSvgRepo) FindByUserID(userID string, limit, offset int) ([]*model.SavedSvg, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, svg_content, created_at, updated_at
		 FROM saved_svgs WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*model.SavedSvg, 0)
	for rows.Next() {
		s := &model.SavedSvg{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.SvgContent, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *SavedSvgRepo) FindByID(id string) (*model.SavedSvg, error) {
	s := &model.SavedSvg{}
	err := r.db.QueryRow(
		`SELECT id, user_id, name, svg_content, created_at, updated_at FROM saved_svgs WHERE id=$1`, id,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.SvgContent, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SavedSvgRepo) Update(id, name, svgContent string) (*model.SavedSvg, error) {
	s := &model.SavedSvg{}
	err := r.db.QueryRow(
		`UPDATE saved_svgs SET name=$1, svg_content=$2, updated_at=$3 WHERE id=$4
		 RETURNING id, user_id, name, svg_content, created_at, updated_at`,
		name, svgContent, time.Now(), id,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.SvgContent, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SavedSvgRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM saved_svgs WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
