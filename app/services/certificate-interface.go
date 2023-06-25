package services

import (
	"context"
	"errors"

	"github.com/fapiko/john-hancock-platform/app/contracts"
)

type CertificateType string

const (
	CertTypeRootCA         CertificateType = "root_ca"
	CertTypeIntermediateCA CertificateType = "intermediate_ca"
	CertTypeCertificate    CertificateType = "certificate"
)

var ErrCertUnautorized = errors.New("user does not have access to this certificate")

func (ct CertificateType) String() string {
	return string(ct)
}

type CertificateService interface {
	DeleteCertForUser(ctx context.Context, id string, userID string) error
	GetCert(ctx context.Context, id string) (*contracts.CertificateResponse, error)
	GetCertAsPEMForUser(ctx context.Context, id string, userID string) (string, error)
	GetUserCerts(
		ctx context.Context,
		userId string,
		certTypes []CertificateType,
	) ([]*contracts.CertificateLightResponse, error)
	CreateCACert(
		ctx context.Context,
		request *contracts.CreateCARequest,
		userID string,
		certType CertificateType,
	) (
		[]byte,
		error,
	)
	CreateCert(
		ctx context.Context,
		caID string,
		request *contracts.CreateCertificateRequest,
		userID string,
	) (*contracts.CertificateLightResponse, error)
	GetCertsByParentCAForUser(
		ctx context.Context,
		parentCA string,
		userID string,
	) ([]*contracts.CertificateLightResponse, error)
}
