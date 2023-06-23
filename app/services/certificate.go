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
	"github.com/fapiko/john-hancock-platform/app/repositories"
)

var _ CertificateService = (*CertificateServiceImpl)(nil)

type CertificateType string

const (
	CertTypeRootCA         CertificateType = "root_ca"
	CertTypeIntermediateCA CertificateType = "intermediate_ca"
	CertTypeServer         CertificateType = "server"
	CertTypeClient         CertificateType = "client"
)

func (ct CertificateType) String() string {
	return string(ct)
}

type CertificateService interface {
	GetCert(ctx context.Context, id string) (*contracts.CertificateResponse, error)
	GetUserCerts(
		ctx context.Context,
		userId string,
		certTypes []CertificateType,
	) ([]*contracts.CertificateLightResponse, error)
	GenerateCert(context.Context, *contracts.CreateCARequest, CertificateType, PrivateKey) (
		[]byte,
		error,
	)
}

type CertificateServiceImpl struct {
	certRepository repositories.CertRepository
}

func NewCertificateServiceImpl(certRepository repositories.CertRepository) *CertificateServiceImpl {
	return &CertificateServiceImpl{
		certRepository: certRepository,
	}
}

type CertInfo struct {
	Cert       *x509.Certificate
	PrivateKey *rsa.PrivateKey
}

func (c *CertificateServiceImpl) GetCert(
	ctx context.Context,
	id string,
) (*contracts.CertificateResponse, error) {
	certDao, err := c.certRepository.GetCertByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDao.Data)
	if err != nil {
		return nil, err
	}

	issuer := &contracts.PkixName{}
	issuer.FromName(&cert.Issuer)

	subject := &contracts.PkixName{}
	subject.FromName(&cert.Subject)

	extKeyUsage := make([]int, 0)
	for _, usage := range cert.ExtKeyUsage {
		extKeyUsage = append(extKeyUsage, int(usage))
	}

	certResponse := &contracts.CertificateResponse{
		ID:                 certDao.ID,
		OwnerID:            certDao.UserID,
		Name:               certDao.Name,
		Type:               certDao.Type,
		Created:            certDao.Created,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		Version:            cert.Version,
		SerialNumber:       int(cert.SerialNumber.Int64()),
		Issuer:             issuer,
		Subject:            subject,
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		IsCA:               cert.IsCA,
		MaxPathLen:         cert.MaxPathLen,
		MaxPathLenZero:     cert.MaxPathLenZero,
		KeyUsage:           int(cert.KeyUsage),
		ExtKeyUsage:        extKeyUsage,
	}

	return certResponse, nil
}

func (c *CertificateServiceImpl) GenerateCert(
	ctx context.Context,
	request *contracts.CreateCARequest,
	certificateType CertificateType,
	key PrivateKey,
) ([]byte, error) {
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

	certData, err := x509.CreateCertificate(
		rand.Reader,
		&certTemplate,
		&certTemplate,
		key.Public(),
		key,
	)
	if err != nil {
		return nil, err
	}

	return certData, nil
}

func (c *CertificateServiceImpl) GetUserCerts(
	ctx context.Context,
	userId string,
	certTypes []CertificateType,
) (
	[]*contracts.CertificateLightResponse,
	error,
) {
	strCertTypes := make([]string, len(certTypes))
	for i, certType := range certTypes {
		strCertTypes[i] = certType.String()
	}

	daos, err := c.certRepository.GetCertsByUserID(ctx, userId, strCertTypes)
	if err != nil {
		return nil, err
	}

	response := make([]*contracts.CertificateLightResponse, len(daos))
	for i, dao := range daos {
		response[i] = &contracts.CertificateLightResponse{
			ID:      dao.ID,
			Name:    dao.Name,
			Type:    dao.Type,
			Created: dao.Created,
		}
	}

	return response, nil
}
