package story

import (
	"encoding/json"
	"time"
)

type Progress struct {
	UserID         string          `db:"user_id" json:"user_id"`
	ChapterID      string          `db:"chapter_id" json:"chapter_id"`
	NodeID         string          `db:"node_id" json:"node_id"`
	ContentVersion string          `db:"content_version" json:"content_version"`
	Flags          json.RawMessage `db:"flags" json:"flags"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}
