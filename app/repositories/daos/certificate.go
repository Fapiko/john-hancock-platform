package daos

import (
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
)

type Certificate struct {
	ID                string `gorm:"type:uuid;primary_key;"`
	UserID            string
	Name              string
	Data              []byte
	Type              string
	Created           time.Time
	ParentCertificate string
	KeyID             string
}

func (d *Certificate) ToLightResponse() *contracts.CertificateLightResponse {
	return &contracts.CertificateLightResponse{
		ID:      d.ID,
		Name:    d.Name,
		Type:    d.Type,
		Created: d.Created,
	}
}
