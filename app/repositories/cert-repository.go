package repositories

import (
	"context"

	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
)

type CertRepository interface {
	CreateCert(
		ctx context.Context,
		userId string,
		name string,
		data []byte,
		certType string,
	) (*daos.Certificate, error)

	GetCertByID(
		ctx context.Context,
		id string,
	) (*daos.Certificate, error)

	GetCertsByUserID(
		ctx context.Context,
		userId string,
		certType string,
	) ([]*daos.Certificate, error)
}
