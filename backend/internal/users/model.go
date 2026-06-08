package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Username     string
	PasswordHash string
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
