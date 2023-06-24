package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ KeyRepository = (*KeyRepositoryMySQL)(nil)

type KeyRepositoryMySQL struct {
	db *gorm.DB
}

func NewKeyRepositoryMySQL(db *gorm.DB) *KeyRepositoryMySQL {
	return &KeyRepositoryMySQL{
		db: db,
	}
}

func (k *KeyRepositoryMySQL) CreateKey(
	ctx context.Context,
	userId string,
	data []byte,
	algorithm string,
	name string,
) (*daos.Key, error) {
	keyDao := &daos.Key{
		ID:        uuid.New().String(),
		UserID:    userId,
		Data:      data,
		Algorithm: algorithm,
		Name:      name,
		Created:   time.Now(),
	}

	result := k.db.WithContext(ctx).Create(keyDao)
	return keyDao, result.Error
}

func (k *KeyRepositoryMySQL) GetKey(ctx context.Context, id string) (*daos.Key, error) {
	keyDao := &daos.Key{}
	result := k.db.WithContext(ctx).Where("id = ?", id).First(keyDao)

	return keyDao, result.Error
}

func (k *KeyRepositoryMySQL) GetKeysForUser(ctx context.Context, userId string) (
	[]*daos.Key,
	error,
) {
	keys := make([]*daos.Key, 0)
	result := k.db.WithContext(ctx).Where("user_id = ?", userId).Find(&keys)

	return keys, result.Error
}
