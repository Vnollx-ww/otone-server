package player

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrPlayerNotFound  = errors.New("player not found")
	ErrInvalidNickname = errors.New("nickname must be 1-32 characters")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Get(ctx context.Context, userID string) (*Player, error) {
	player, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, ErrPlayerNotFound
	}
	return player, nil
}

func (s *Service) UpdateNickname(ctx context.Context, userID, nickname string) (*Player, error) {
	nickname = strings.TrimSpace(nickname)
	if n := utf8.RuneCountInString(nickname); n < 1 || n > 32 {
		return nil, ErrInvalidNickname
	}
	if err := s.repo.UpdateNickname(ctx, userID, nickname); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}
