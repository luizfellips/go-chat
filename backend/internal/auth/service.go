package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/users"
)

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Tokens TokenPair
	User   *users.User
}

type Service struct {
	users         users.Repository
	refreshTokens RefreshTokenRepository
	tokens        *TokenService
}

func NewService(users users.Repository, refreshTokens RefreshTokenRepository, tokens *TokenService) *Service {
	return &Service{users: users, refreshTokens: refreshTokens, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*users.User, error) {
	hash, err := s.tokens.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.users.Create(ctx, strings.ToLower(input.Email), input.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, apperr.ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(input.Email))
	if err != nil {
		return nil, apperr.ErrInvalidCredentials
	}
	if !s.tokens.CheckPassword(user.PasswordHash, input.Password) {
		return nil, apperr.ErrInvalidCredentials
	}
	if err := s.refreshTokens.RevokeAllForUser(ctx, user.ID); err != nil {
		return nil, err
	}
	pair, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Tokens: *pair, User: user}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginOutput, error) {
	hash := HashToken(refreshToken)
	raw, newHash, expiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	userID, err := s.refreshTokens.RotateRefreshToken(ctx, hash, newHash, expiresAt)
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	access, expiresIn, err := s.tokens.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{
		Tokens: TokenPair{AccessToken: access, RefreshToken: raw, ExpiresIn: expiresIn},
		User:   user,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := HashToken(refreshToken)
	return s.refreshTokens.Revoke(ctx, hash)
}

func (s *Service) issueTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	access, expiresIn, err := s.tokens.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	raw, hash, expiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.refreshTokens.Create(ctx, userID, hash, expiresAt); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: raw, ExpiresIn: expiresIn}, nil
}
