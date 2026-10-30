package story

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

// ListByUser returns the player's saved progress for every chapter. Chapters
// are a bounded domain set, so LIMIT 200 is a safety cap rather than pagination.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Progress, error) {
	var items []Progress
	err := r.db.SelectContext(ctx, &items, `
		SELECT user_id, chapter_id, node_id, content_version, flags, updated_at
		FROM story_progress
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT 200`, userID)
	if err != nil {
		return nil, fmt.Errorf("list story progress: %w", err)
	}
	return items, nil
}

func (r *Repository) Get(ctx context.Context, userID, chapterID string) (*Progress, error) {
	var progress Progress
	err := r.db.GetContext(ctx, &progress, `
		SELECT user_id, chapter_id, node_id, content_version, flags, updated_at
		FROM story_progress
		WHERE user_id = ? AND chapter_id = ?`, userID, chapterID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get story progress: %w", err)
	}
	return &progress, nil
}

func (r *Repository) Upsert(ctx context.Context, progress *Progress) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO story_progress (user_id, chapter_id, node_id, content_version, flags)
		VALUES (?, ?, ?, ?, ?) AS vals
		ON DUPLICATE KEY UPDATE
			node_id = vals.node_id,
			content_version = vals.content_version,
			flags = vals.flags,
			updated_at = UTC_TIMESTAMP(3)`,
		progress.UserID, progress.ChapterID, progress.NodeID,
		progress.ContentVersion, string(progress.Flags))
	if err != nil {
		return fmt.Errorf("upsert story progress: %w", err)
	}
	return nil
}
