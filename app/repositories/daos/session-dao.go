package daos

import (
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
)

type Session struct {
	ID         string `gorm:"type:uuid;primary_key;"`
	Created    time.Time
	Expiration time.Time
	UserID     string `gorm:"type:uuid;"`
}

func (s *Session) ToResponse() *contracts.SessionResponse {
	return &contracts.SessionResponse{
		ID:        s.ID,
		CreatedAt: s.Created,
		Expires:   s.Expiration,
	}
}
