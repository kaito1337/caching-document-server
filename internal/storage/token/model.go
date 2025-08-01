package storage

import (
	"time"

	"github.com/google/uuid"
)

type UserToken struct {
	Token     string    `db:"token"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
}
