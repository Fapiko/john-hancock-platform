package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
)

const (
	paramSessionID = "sessionID"
)

var bcryptCost = 14
var sessionExpiration = 24 * time.Hour

type UserRepository interface {
	CleanupSessions(ctx context.Context) (int, error)
	CreateSession(ctx context.Context, userID string) (*contracts.SessionResponse, error)
	CreateUser(ctx context.Context, user *contracts.CreateUserRequest) (*daos.User, error)
	GetUserByEmail(ctx context.Context, email string) (*daos.User, error)
	GetUserBySessionID(ctx context.Context, sessionID string) (*daos.User, error)
}
