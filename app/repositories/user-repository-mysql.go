package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"gorm.io/gorm"
)

var _ UserRepository = (*UserRepositoryMySql)(nil)

type UserRepositoryMySql struct {
	db *gorm.DB
}

func NewUserRepositoryMySql(db *gorm.DB) *UserRepositoryMySql {
	return &UserRepositoryMySql{
		db: db,
	}
}

func (u *UserRepositoryMySql) CleanupSessions(ctx context.Context) (int, error) {
	result := u.db.Where("expiration <= ?", time.Now()).Delete(&daos.Session{})
	return int(result.RowsAffected), result.Error
}

func (u *UserRepositoryMySql) CreateSession(
	ctx context.Context,
	userID string,
) (*contracts.SessionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryMySql) CreateUser(
	ctx context.Context,
	user *contracts.CreateUserRequest,
) (*daos.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryMySql) GetUserByEmail(ctx context.Context, email string) (
	*daos.User,
	error,
) {
	user := &daos.User{}
	result := u.db.WithContext(ctx).Where("email = ?", email).First(user)

	return user, result.Error
}

func (u *UserRepositoryMySql) GetUserBySessionID(ctx context.Context, sessionID string) (
	*daos.User,
	error,
) {
	//TODO implement me
	panic("implement me")
}
