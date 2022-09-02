package contracts

import "time"

type SessionResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Expires   time.Time `json:"expires"`
}
