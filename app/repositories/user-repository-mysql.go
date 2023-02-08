package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	sessionDao := &daos.Session{
		ID:         uuid.New().String(),
		Created:    time.Now(),
		Expiration: time.Now().Add(sessionExpiration),
		UserID:     userID,
	}

	result := u.db.WithContext(ctx).Create(sessionDao)
	if result.Error != nil {
		return nil, result.Error
	}

	return sessionDao.ToResponse(), nil
}

func (u *UserRepositoryMySql) CreateUser(
	ctx context.Context,
	createUser *contracts.CreateUserRequest,
) (*daos.User, error) {
	passwordHashData, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcryptCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(passwordHashData)
	userID := uuid.New().String()

	user := &daos.User{
		ID:        userID,
		FirstName: createUser.FirstName,
		LastName:  createUser.LastName,
		Email:     createUser.Email,
		Password:  passwordHash,
	}

	result := u.db.WithContext(ctx).Create(user)

	return user, result.Error
}

func (u *UserRepositoryMySql) GetUserByEmail(ctx context.Context, email string) (
	*daos.User,
	error,
) {
	user := &daos.User{}
	result := u.db.WithContext(ctx).Where("email = ?", email).First(user)

	return user, convertNotFound(result.Error)
}

func (u *UserRepositoryMySql) GetUserBySessionID(ctx context.Context, sessionID string) (
	*daos.User,
	error,
) {
	// Fetch session by ID
	session := &daos.Session{}
	result := u.db.WithContext(ctx).Where("id = ?", sessionID).First(session)
	if result.Error != nil {
		return nil, convertNotFound(result.Error)
	}

	// Fetch user by ID
	user := &daos.User{}
	result = u.db.WithContext(ctx).Where("id = ?", session.UserID).First(user)

	return user, convertNotFound(result.Error)
}
