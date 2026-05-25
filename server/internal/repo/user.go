package repo

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) FindByProvider(provider, providerID string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE provider=$1 AND provider_id=$2`,
		provider, providerID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) Create(u *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO users (email, name, avatar_url, provider, provider_id)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`,
		u.Email, u.Name, u.AvatarURL, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) UpsertByProvider(u *model.User) error {
	email := u.Email
	if email == "" {
		email = ""
	}
	return r.db.QueryRow(
		`INSERT INTO users (email, name, avatar_url, provider, provider_id)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (provider, provider_id) DO UPDATE
		 SET name=EXCLUDED.name, avatar_url=EXCLUDED.avatar_url, email=COALESCE(NULLIF(EXCLUDED.email,''), users.email), updated_at=now()
		 RETURNING id, email, created_at, updated_at`,
		nullIfEmpty(u.Email), u.Name, u.AvatarURL, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *UserRepo) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) CreateGuest() (*model.User, error) {
	guestID := randomHex(16)
	u := &model.User{
		Name:       "游客",
		Provider:   "guest",
		ProviderID: guestID,
	}
	err := r.db.QueryRow(
		`INSERT INTO users (name, provider, provider_id) VALUES ($1,$2,$3)
		 RETURNING id, email, created_at, updated_at`,
		u.Name, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) FindByGuestID(guestID string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE provider='guest' AND provider_id=$1`, guestID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) SaveVerificationCode(email, code string) error {
	_, err := r.db.Exec(
		`INSERT INTO verification_codes (email, code, expires_at) VALUES ($1,$2,$3)`,
		email, code, time.Now().Add(5*time.Minute),
	)
	return err
}

func (r *UserRepo) VerifyCode(email, code string) (bool, error) {
	var used bool
	err := r.db.QueryRow(
		`SELECT used FROM verification_codes
		 WHERE email=$1 AND code=$2 AND expires_at > now()
		 ORDER BY created_at DESC LIMIT 1`,
		email, code,
	).Scan(&used)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if used {
		return false, nil
	}
	_, err = r.db.Exec(`UPDATE verification_codes SET used=true WHERE email=$1 AND code=$2`, email, code)
	return !used, err
}

func (r *UserRepo) LastCodeSentAt(email string) (time.Time, error) {
	var t time.Time
	err := r.db.QueryRow(
		`SELECT created_at FROM verification_codes
		 WHERE email=$1 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&t)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return t, err
}

func (r *UserRepo) CleanupExpiredCodes() error {
	_, err := r.db.Exec(`DELETE FROM verification_codes WHERE expires_at < now() - INTERVAL '1 hour'`)
	return err
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
