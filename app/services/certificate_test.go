package services

import (
	"context"
	"testing"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestCertificateServiceImpl_GenerateCert(t *testing.T) {
	ctx := context.Background()
	service := NewCertificateServiceImpl(nil)
	_, err := service.GenerateCert(ctx, &contracts.CreateCARequest{}, "", CertTypeRootCA)

	//block, _ := pem.Decode(cert)
	//t.Log(block.Type)

	assert.NoError(t, err)

	//cert := pem.Decode(certInfo)
	//
	//data := pem.Block{Type: "CERTIFICATE", Bytes: cert.Bytes}
	//certPEM := pem.EncodeToMemory(&data)
	//
	//privData := pem.Block{"RSA PRIVATE KEY", nil, x509.MarshalPKCS1PrivateKey(certInfo.PrivateKey)}
	//privPEM := pem.EncodeToMemory(&privData)
	//
	//t.Log(string(certPEM))
	//t.Log(string(privPEM))
}
