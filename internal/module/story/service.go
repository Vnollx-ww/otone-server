package story

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrProgressNotFound = errors.New("story progress not found")
	ErrInvalidProgress  = errors.New("chapter_id and node_id are required")
	ErrInvalidFlags     = errors.New("flags must be a JSON object and at most 16KB")
)

const maxFlagsBytes = 16 * 1024

type Store interface {
	ListByUser(ctx context.Context, userID string) ([]Progress, error)
	Get(ctx context.Context, userID, chapterID string) (*Progress, error)
	Upsert(ctx context.Context, progress *Progress) error
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

type SaveInput struct {
	ChapterID      string
	NodeID         string
	ContentVersion string
	Flags          json.RawMessage
}

func (s *Service) List(ctx context.Context, userID string) ([]Progress, error) {
	items, err := s.store.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []Progress{}
	}
	return items, nil
}

func (s *Service) Save(ctx context.Context, userID string, in SaveInput) (*Progress, error) {
	chapterID := strings.TrimSpace(in.ChapterID)
	nodeID := strings.TrimSpace(in.NodeID)
	contentVersion := strings.TrimSpace(in.ContentVersion)
	if chapterID == "" || nodeID == "" {
		return nil, ErrInvalidProgress
	}

	flags := bytes.TrimSpace(in.Flags)
	if len(flags) == 0 {
		flags = json.RawMessage(`{}`)
	}
	if len(flags) > maxFlagsBytes || flags[0] != '{' || !json.Valid(flags) {
		return nil, ErrInvalidFlags
	}

	progress := &Progress{
		UserID:         userID,
		ChapterID:      chapterID,
		NodeID:         nodeID,
		ContentVersion: contentVersion,
		Flags:          json.RawMessage(flags),
		UpdatedAt:      s.now().UTC(),
	}
	if err := s.store.Upsert(ctx, progress); err != nil {
		return nil, err
	}

	saved, err := s.store.Get(ctx, userID, chapterID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrProgressNotFound
	}
	return saved, nil
}
