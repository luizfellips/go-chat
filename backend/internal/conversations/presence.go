package conversations

import "github.com/google/uuid"

type PresenceChecker interface {
	IsOnline(userID uuid.UUID) bool
}
