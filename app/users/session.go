package users

import "time"

type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Expires   time.Time `json:"expires"`
}
