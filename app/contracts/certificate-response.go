package contracts

import "time"

type CertificateResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Created time.Time `json:"created"`
}
