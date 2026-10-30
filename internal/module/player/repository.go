package player

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Get(ctx context.Context, userID string) (*Player, error) {
	var player Player
	err := r.db.GetContext(ctx, &player, `
		SELECT user_id, nickname, level, exp, created_at, updated_at
		FROM players
		WHERE user_id = ?`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get player: %w", err)
	}
	return &player, nil
}

func (r *Repository) UpdateNickname(ctx context.Context, userID, nickname string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE players SET nickname = ? WHERE user_id = ?`, nickname, userID)
	if err != nil {
		return fmt.Errorf("update player nickname: %w", err)
	}
	return nil
}
