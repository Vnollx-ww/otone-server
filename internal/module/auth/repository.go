package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindByDevice(ctx context.Context, deviceID, platform string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, device_id, platform, status, created_at, updated_at
		FROM users
		WHERE device_id = ? AND platform = ?
		LIMIT 1`, deviceID, platform)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by device: %w", err)
	}
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, device_id, platform, status)
		VALUES (?, ?, ?, ?)`, user.ID, user.DeviceID, user.Platform, user.Status); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO players (user_id) VALUES (?)`, user.ID); err != nil {
		return fmt.Errorf("insert player: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create user tx: %w", err)
	}
	return nil
}

func IsDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *Repository) FindRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.GetContext(ctx, &token, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = ?
		LIMIT 1`, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &token, nil
}

// RotateRefreshToken revokes the current token and persists its replacement in
// a single transaction. Revoking an already-used token affects zero rows, which
// makes concurrent refresh attempts fail safely.
func (r *Repository) RotateRefreshToken(ctx context.Context, currentID string, next *RefreshToken) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rotate token tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND revoked_at IS NULL`, currentID)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke refresh token rows affected: %w", err)
	}
	if affected == 0 {
		return ErrInvalidRefreshToken
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`, next.ID, next.UserID, next.TokenHash, next.ExpiresAt); err != nil {
		return fmt.Errorf("insert rotated refresh token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit rotate token tx: %w", err)
	}
	return nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, hash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = UTC_TIMESTAMP(3)
		WHERE token_hash = ? AND revoked_at IS NULL`, hash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}
