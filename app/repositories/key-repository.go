package repositories

import (
	"context"

	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
)

type KeyRepository interface {
	CreateKey(
		ctx context.Context,
		userId string,
		data []byte,
		algorithm string,
		name string,
	) (*daos.Key, error)
	GetKey(ctx context.Context, id string) (*daos.Key, error)
	GetKeysForUser(
		ctx context.Context,
		userId string,
	) ([]*daos.Key, error)
}
