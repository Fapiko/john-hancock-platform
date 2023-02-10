package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
)

var _ CertificateService = (*CertificateServiceImpl)(nil)

type CertificateType string

const (
	CertTypeRootCA         CertificateType = "root_ca"
	CertTypeIntermediateCA                 = "intermediate_ca"
	CertTypeServer                         = "server"
	CertTypeClient                         = "client"
)

func (ct CertificateType) String() string {
	return string(ct)
}

type CertificateService interface {
	GenerateCert(context.Context, *contracts.CreateCARequest, CertificateType) ([]byte, error)
}

type CertificateServiceImpl struct {
}

func NewCertificateServiceImpl() *CertificateServiceImpl {
	return &CertificateServiceImpl{}
}

type CertInfo struct {
	Cert       *x509.Certificate
	PrivateKey *rsa.PrivateKey
}

func (c *CertificateServiceImpl) GenerateCert(
	ctx context.Context,
	request *contracts.CreateCARequest,
	certificateType CertificateType,
) (
	[]byte,
	error,
) {
	var maxPathLen = 0
	if certificateType == CertTypeRootCA {
		maxPathLen = 1 // TODO: Make this configurable - may want multiple intermediate CAs
	}

	var keyUsage = x509.KeyUsage(0)
	var isCA = false
	if certificateType == CertTypeRootCA || certificateType == CertTypeIntermediateCA {
		isCA = true
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:       []string{request.Country},
			Organization:  []string{request.Organization},
			Locality:      []string{request.Locality},
			Province:      []string{request.State},
			StreetAddress: []string{request.StreetAddress},
			PostalCode:    []string{request.PostalCode},
			CommonName:    request.Name,
		},
		NotBefore:             time.Now(),
		NotAfter:              request.Expiration,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny}, // TODO: Make this configurable
		IsCA:                  isCA,
		MaxPathLen:            maxPathLen,
		MaxPathLenZero:        true,
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 4096) // TODO: Make private key size configurable
	if err != nil {
		return nil, err
	}

	certData, err := x509.CreateCertificate(
		rand.Reader,
		&certTemplate,
		&certTemplate,
		&priv.PublicKey,
		priv,
	)
	if err != nil {
		return nil, err
	}

	return certData, nil
}
