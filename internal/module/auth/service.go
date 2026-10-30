package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserDisabled        = errors.New("user is disabled")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidDevice       = errors.New("device_id is required")
)

type Service struct {
	repo   *Repository
	tokens *TokenManager
	now    func() time.Time
}

func NewService(repo *Repository, tokens *TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens, now: time.Now}
}

// GuestLogin finds or creates the account bound to a device and issues a token
// pair. The unique key on (device_id, platform) makes this safe under
// concurrent first launches.
func (s *Service) GuestLogin(ctx context.Context, deviceID, platform string) (*User, *TokenPair, error) {
	deviceID = strings.TrimSpace(deviceID)
	platform = strings.TrimSpace(platform)
	if deviceID == "" || len(deviceID) > 128 {
		return nil, nil, ErrInvalidDevice
	}

	user, err := s.repo.FindByDevice(ctx, deviceID, platform)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		user = &User{ID: uuid.NewString(), DeviceID: deviceID, Platform: platform, Status: 1}
		if err := s.repo.Create(ctx, user); err != nil && !IsDuplicateKey(err) {
			return nil, nil, err
		}
		// Reload to pick up database-generated timestamps, and to return the
		// winner's row when a concurrent request created the same device first.
		user, err = s.repo.FindByDevice(ctx, deviceID, platform)
		if err != nil {
			return nil, nil, err
		}
		if user == nil {
			return nil, nil, fmt.Errorf("create user: row not found after insert")
		}
	}
	if user.Status != 1 {
		return nil, nil, ErrUserDisabled
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*TokenPair, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	current, err := s.repo.FindRefreshTokenByHash(ctx, HashRefreshToken(rawToken))
	if err != nil {
		return nil, err
	}
	if current == nil || current.RevokedAt != nil || !s.now().UTC().Before(current.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	accessToken, err := s.tokens.NewAccessToken(current.UserID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	newRaw, newHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	next := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    current.UserID,
		TokenHash: newHash,
		ExpiresAt: now.Add(s.tokens.RefreshTTL()),
	}
	if err := s.repo.RotateRefreshToken(ctx, current.ID, next); err != nil {
		return nil, err
	}
	return s.tokenPair(accessToken, newRaw, now), nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil
	}
	return s.repo.RevokeRefreshToken(ctx, HashRefreshToken(rawToken))
}

func (s *Service) issueTokens(ctx context.Context, userID string) (*TokenPair, error) {
	accessToken, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	rawRefresh, refreshHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	token := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(s.tokens.RefreshTTL()),
	}
	if err := s.repo.CreateRefreshToken(ctx, token); err != nil {
		return nil, err
	}
	return s.tokenPair(accessToken, rawRefresh, now), nil
}

func (s *Service) tokenPair(accessToken, refreshToken string, now time.Time) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokens.AccessTTL().Seconds()),
		ExpiresAt:    now.Add(s.tokens.AccessTTL()),
	}
}
