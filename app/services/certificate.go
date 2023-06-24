package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories"
)

var _ CertificateService = (*CertificateServiceImpl)(nil)

type CertificateServiceImpl struct {
	certRepository repositories.CertRepository
	keyRepository  repositories.KeyRepository
	keyService     KeyService
}

func (c *CertificateServiceImpl) GetCertAsPEMForUser(
	ctx context.Context,
	id string,
	userID string,
) (string, error) {
	cert, err := c.certRepository.GetCertByID(ctx, id)
	if err != nil {
		return "", err
	}

	if cert.UserID != userID {
		return "", errors.New("user does not have access to this certificate")
	}

	pb := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Data,
	}

	return string(pem.EncodeToMemory(pb)), nil
}

func (c *CertificateServiceImpl) GetCertsByParentCAForUser(
	ctx context.Context,
	parentCA string,
	userID string,
) ([]*contracts.CertificateLightResponse, error) {
	certDaos, err := c.certRepository.GetCertsByParentCA(ctx, parentCA)
	if err != nil {
		return nil, err
	}

	var certResponses []*contracts.CertificateLightResponse
	for _, cert := range certDaos {
		if cert.UserID != userID {
			continue
		}

		certResponses = append(certResponses, cert.ToLightResponse())
	}

	return certResponses, nil
}

func (c *CertificateServiceImpl) CreateCert(
	ctx context.Context,
	caID string,
	request *contracts.CreateCertificateRequest,
	userID string,
) (*contracts.CertificateLightResponse, error) {
	caCert, err := c.getX509CertificateForUser(ctx, caID, userID)
	if err != nil {
		return nil, err
	}

	caKeyId, err := c.certRepository.GetKeyIDByCertID(ctx, caID)
	if err != nil {
		return nil, err
	}
	caPrivateKey, err := c.keyService.GetDecryptedKeyForUser(
		ctx,
		caKeyId,
		userID,
		request.CAKeyPassword,
	)
	if err != nil {
		return nil, err
	}

	certKey, err := c.keyService.GetDecryptedKeyForUser(
		ctx,
		request.KeyId,
		userID,
		request.KeyPassword,
	)
	if err != nil {
		return nil, err
	}

	keyUsage, extKeyusage, err := c.keyUsages(request.KeyUsages)

	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: request.CommonName,
		},
		DNSNames:              request.SubjectAlternativeNames,
		NotBefore:             time.Now(),
		NotAfter:              request.Expiration,
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyusage,
	}

	cert, err := x509.CreateCertificate(
		rand.Reader,
		&certTemplate,
		caCert,
		certKey.Public(),
		caPrivateKey,
	)
	if err != nil {
		return nil, err
	}

	dao, err := c.certRepository.CreateCert(
		ctx,
		userID,
		request.Name,
		cert,
		CertTypeCertificate.String(),
		caID,
		request.KeyId,
	)
	if err != nil {
		return nil, err
	}

	return &contracts.CertificateLightResponse{
		ID:      dao.ID,
		Name:    dao.Name,
		Type:    dao.Type,
		Created: dao.Created,
	}, nil
}

func (c *CertificateServiceImpl) keyUsages(keyUsages []string) (
	x509.KeyUsage,
	[]x509.ExtKeyUsage,
	error,
) {
	keyUsage := x509.KeyUsage(0)
	extKeyUsage := make([]x509.ExtKeyUsage, 0)

	for _, usage := range keyUsages {
		switch usage {
		case "digitalSignature":
			keyUsage |= x509.KeyUsageDigitalSignature
		case "contentCommitment":
			keyUsage |= x509.KeyUsageContentCommitment
		case "keyEncipherment":
			keyUsage |= x509.KeyUsageKeyEncipherment
		case "dataEncipherment":
			keyUsage |= x509.KeyUsageDataEncipherment
		case "keyAgreement":
			keyUsage |= x509.KeyUsageKeyAgreement
		case "certSign":
			keyUsage |= x509.KeyUsageCertSign
		case "crlSign":
			keyUsage |= x509.KeyUsageCRLSign
		case "encipherOnly":
			keyUsage |= x509.KeyUsageEncipherOnly
		case "decipherOnly":
			keyUsage |= x509.KeyUsageDecipherOnly
		case "serverAuth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageServerAuth)
		case "clientAuth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageClientAuth)
		case "codeSigning":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageCodeSigning)
		case "emailProtection":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageEmailProtection)
		case "ipsec_end_system":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECEndSystem)
		case "ipsec_tunnel":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECTunnel)
		case "ipsec_user":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECUser)
		case "timeStamping":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageTimeStamping)
		case "ocspSigning":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageOCSPSigning)
		default:
			return 0, nil, errors.New("invalid key usage")
		}
	}

	return keyUsage, extKeyUsage, nil
}

func (c *CertificateServiceImpl) getX509CertificateForUser(
	ctx context.Context,
	id string,
	userId string,
) (*x509.Certificate, error) {
	ca, err := c.certRepository.GetCertByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ca.UserID != userId {
		return nil, errors.New("certificate authority does not belong to user")
	}

	return x509.ParseCertificate(ca.Data)
}

func NewCertificateServiceImpl(
	certRepository repositories.CertRepository,
	keyRepository repositories.KeyRepository,
	keyService KeyService,
) *CertificateServiceImpl {
	return &CertificateServiceImpl{
		certRepository: certRepository,
		keyRepository:  keyRepository,
		keyService:     keyService,
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
		KeyID:              certDao.KeyID,
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

func (c *CertificateServiceImpl) CreateCACert(
	ctx context.Context,
	request *contracts.CreateCARequest,
	userID string,
	certificateType CertificateType,
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

	key, err := c.keyService.GetDecryptedKeyForUser(ctx, request.KeyID, userID, request.KeyPassword)
	if err != nil {
		return nil, err
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

	var parentCert *x509.Certificate
	var parentKey PrivateKey
	if request.ParentCA == "" {
		parentCert = &certTemplate
		parentKey = key
	} else {
		parentCert, err = c.getX509CertificateForUser(ctx, request.ParentCA, userID)
		if err != nil {
			return nil, err
		}

		keyId, err := c.certRepository.GetKeyIDByCertID(ctx, request.ParentCA)
		if err != nil {
			return nil, err
		}

		parentKey, err = c.keyService.GetDecryptedKeyForUser(
			ctx,
			keyId,
			userID,
			request.ParentKeyPassword,
		)
		if err != nil {
			return nil, err
		}
	}

	certData, err := x509.CreateCertificate(
		rand.Reader,
		&certTemplate,
		parentCert,
		key.Public(),
		parentKey,
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
