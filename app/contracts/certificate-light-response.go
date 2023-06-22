package contracts

import "time"

type CertificateLightResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Created time.Time `json:"created"`
}
