package services

import (
	"testing"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestKeyServiceImpl_CreateKey(t *testing.T) {
	keyService := NewKeyServiceImpl(nil)
	data, err := keyService.CreateKey(nil, contracts.ED25519, "pwd123")

	assert.NoError(t, err)
	t.Log(string(data))
}
