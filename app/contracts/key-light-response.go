package contracts

import "time"

type KeyLightResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Created   time.Time `json:"created"`
	Algorithm string    `json:"algorithm"`
}
