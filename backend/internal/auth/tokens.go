package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	accessSecret []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

func NewTokenService(accessSecret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		accessSecret: []byte(accessSecret),
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
	}
}

type accessClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"sub"`
}

func (s *TokenService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *TokenService) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *TokenService) GenerateAccessToken(userID uuid.UUID) (string, int64, error) {
	now := time.Now()
	exp := now.Add(s.accessTTL)
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(s.accessTTL.Seconds()), nil
}

func (s *TokenService) ParseAccessToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}

func (s *TokenService) GenerateRefreshToken() (string, string, time.Time, error) {
	raw := uuid.New().String()
	hash := HashToken(raw)
	expiresAt := time.Now().Add(s.refreshTTL)
	return raw, hash, expiresAt, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *TokenService) RefreshTTL() time.Duration {
	return s.refreshTTL
}
