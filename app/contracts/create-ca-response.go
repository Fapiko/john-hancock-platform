package contracts

import "time"

type CreateCAResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Type    string    `json:"type"`
}
