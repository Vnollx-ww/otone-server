package player

import "time"

type Player struct {
	UserID    string    `db:"user_id" json:"user_id"`
	Nickname  string    `db:"nickname" json:"nickname"`
	Level     int       `db:"level" json:"level"`
	Exp       int64     `db:"exp" json:"exp"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
