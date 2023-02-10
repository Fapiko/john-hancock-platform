package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ CertRepository = (*CertRepositoryMySQL)(nil)

type CertRepositoryMySQL struct {
	db *gorm.DB
}

func NewCertRepositoryMySQL(db *gorm.DB) *CertRepositoryMySQL {
	return &CertRepositoryMySQL{
		db: db,
	}
}

func (c *CertRepositoryMySQL) CreateCert(
	ctx context.Context,
	userId string,
	name string,
	data []byte,
	certType string,
) (*daos.Certificate, error) {
	certDao := &daos.Certificate{
		ID:      uuid.New().String(),
		Name:    name,
		Data:    data,
		Type:    certType,
		UserID:  userId,
		Created: time.Now(),
	}

	result := c.db.WithContext(ctx).Create(certDao)
	return certDao, result.Error
}