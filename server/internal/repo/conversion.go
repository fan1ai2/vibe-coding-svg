package repo

import (
	"database/sql"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type ConversionRepo struct{ db *sql.DB }

func NewConversionRepo(db *sql.DB) *ConversionRepo { return &ConversionRepo{db} }

func (r *ConversionRepo) Create(c *model.Conversion) error {
	return r.db.QueryRow(
		`INSERT INTO conversions (user_id, status, original_url, format_in, file_size_in)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		c.UserID, c.Status, c.OriginalURL, c.FormatIn, c.FileSizeIn,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *ConversionRepo) FindByID(id string) (*model.Conversion, error) {
	c := &model.Conversion{}
	err := r.db.QueryRow(
		`SELECT id, user_id, status, original_url, svg_url, thumbnail_url,
		 file_size_in, file_size_out, path_count, color_count, format_in,
		 error_message, created_at, completed_at FROM conversions WHERE id=$1`, id,
	).Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
		&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
		&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ConversionRepo) FindByUserID(userID string, limit, offset int) ([]*model.Conversion, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, status, original_url, svg_url, thumbnail_url,
		 file_size_in, file_size_out, path_count, color_count, format_in,
		 error_message, created_at, completed_at
		 FROM conversions WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*model.Conversion, 0)
	for rows.Next() {
		c := &model.Conversion{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
			&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
			&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *ConversionRepo) UpdateStatus(id, status, errMsg string) error {
	var completedAt *time.Time
	if status == model.StatusCompleted || status == model.StatusFailed {
		now := time.Now()
		completedAt = &now
	}
	res, err := r.db.Exec(
		`UPDATE conversions SET status=$1, error_message=$2, completed_at=$3 WHERE id=$4`,
		status, errMsg, completedAt, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ConversionRepo) UpdateResult(id, svgURL, thumbnailURL string, fileSizeOut, pathCount, colorCount int) error {
	now := time.Now()
	res, err := r.db.Exec(
		`UPDATE conversions SET status=$1, svg_url=$2, thumbnail_url=$3,
		 file_size_out=$4, path_count=$5, color_count=$6, completed_at=$7 WHERE id=$8`,
		model.StatusCompleted, svgURL, thumbnailURL, fileSizeOut, pathCount, colorCount, now, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ConversionRepo) GetTodayQuota(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT count FROM daily_quotas WHERE user_id=$1 AND date=CURRENT_DATE`, userID,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func (r *ConversionRepo) IncrementQuota(userID string) error {
	_, err := r.db.Exec(
		`INSERT INTO daily_quotas (user_id, date, count) VALUES ($1, CURRENT_DATE, 1)
		 ON CONFLICT (user_id, date) DO UPDATE SET count = daily_quotas.count + 1`,
		userID,
	)
	return err
}
