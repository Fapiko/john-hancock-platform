package services

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"go.step.sm/crypto/pemutil"
	"golang.org/x/crypto/ed25519"
)

var _ KeyService = (*KeyServiceImpl)(nil)

// PrivateKey is a  custom interface - all crypto packages implement this interface, but
// crypto.PrivateKey type is any for backwards compat
type PrivateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

type KeyService interface {
	CreateKey(
		ctx context.Context,
		userId string,
		name string,
		algorithm contracts.KeyAlgorithm,
		password string,
	) (
		*contracts.KeyLightResponse,
		error,
	)
	GetKeysForUser(
		ctx context.Context,
		userId string,
	) ([]*contracts.KeyLightResponse, error)
}

type KeyServiceImpl struct {
	keyRepository repositories.KeyRepository
}

func NewKeyServiceImpl(keyRepository repositories.KeyRepository) *KeyServiceImpl {
	return &KeyServiceImpl{
		keyRepository: keyRepository,
	}
}

func (k *KeyServiceImpl) CreateKey(
	ctx context.Context,
	userId string,
	name string,
	algorithm contracts.KeyAlgorithm,
	password string,
) (*contracts.KeyLightResponse, error) {
	var privKey any
	var err error

	switch algorithm {
	case contracts.RSA:
		privKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case contracts.ECDSA:
		privKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case contracts.ED25519:
		_, privKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		return nil, errors.New("unsupported algorithm")
	}

	if err != nil {
		return nil, err
	}

	data, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	var pemData []byte
	if password == "" {
		pemData = pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: data,
			},
		)
	} else {
		encrypted, err := pemutil.EncryptPKCS8PrivateKey(
			rand.Reader,
			data,
			[]byte(password),
			x509.PEMCipherAES256,
		)
		if err != nil {
			return nil, err
		}

		pemData = pem.EncodeToMemory(encrypted)
	}

	dao, err := k.keyRepository.CreateKey(ctx, userId, pemData, algorithm.String(), name)
	if err != nil {
		return nil, err
	}

	return &contracts.KeyLightResponse{
		ID:        dao.ID,
		Name:      dao.Name,
		Created:   dao.Created,
		Algorithm: dao.Algorithm,
	}, nil
}

func (k *KeyServiceImpl) GetKeysForUser(
	ctx context.Context,
	userId string,
) ([]*contracts.KeyLightResponse, error) {
	daos, err := k.keyRepository.GetKeysForUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	keys := make([]*contracts.KeyLightResponse, len(daos))
	for i, dao := range daos {
		keys[i] = &contracts.KeyLightResponse{
			ID:        dao.ID,
			Name:      dao.Name,
			Created:   dao.Created,
			Algorithm: dao.Algorithm,
		}
	}

	return keys, nil
}
