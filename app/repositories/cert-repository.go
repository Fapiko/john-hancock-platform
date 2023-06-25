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
		parentCA string,
		keyId string,
	) (*daos.Certificate, error)

	DeleteCertByID(
		ctx context.Context,
		id string,
	) error

	GetCertByID(
		ctx context.Context,
		id string,
	) (*daos.Certificate, error)

	GetCertsByUserID(
		ctx context.Context,
		userId string,
		certTypes []string,
	) ([]*daos.Certificate, error)

	GetKeyIDByCertID(
		ctx context.Context,
		certID string,
	) (string, error)

	GetCertsByParentCA(
		ctx context.Context,
		parentCA string,
	) ([]*daos.Certificate, error)
}
