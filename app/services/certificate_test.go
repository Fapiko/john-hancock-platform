package services

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertificateServiceImpl_GenerateCert(t *testing.T) {
	service := NewCertificateServiceImpl()
	certInfo, err := service.GenerateCert()

	assert.NoError(t, err)

	data := pem.Block{Type: "CERTIFICATE", Bytes: certInfo.Cert.Raw}
	certPEM := pem.EncodeToMemory(&data)

	privData := pem.Block{"RSA PRIVATE KEY", nil, x509.MarshalPKCS1PrivateKey(certInfo.PrivateKey)}
	privPEM := pem.EncodeToMemory(&privData)

	t.Log(string(certPEM))
	t.Log(string(privPEM))
}
