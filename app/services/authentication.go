package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"github.com/fapiko/john-hancock-platform/app/utils"
	"google.golang.org/api/idtoken"
)

var _ AuthService = (*AuthServiceImpl)(nil)

type OAuthClaims struct {
	Email     string
	FirstName string
	LastName  string
}

type AuthService interface {
	GetUserForRequest(ctx context.Context, r *http.Request) (*daos.User, error)
	ValidateOAuthToken(
		ctx context.Context,
		provider string,
		accessToken string,
	) (*daos.User, error)
}

type AuthServiceImpl struct {
	userRepository repositories.UserRepository
}

func (s *AuthServiceImpl) GetUserForRequest(ctx context.Context, r *http.Request) (
	*daos.User,
	error,
) {
	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		return nil, ErrUnauthorized
	}

	user, err := s.userRepository.GetUserBySessionID(ctx, sessionID)
	if err != nil {
		if err == repositories.ErrNoRecord {
			return nil, ErrUnauthorized
		}

		return nil, err
	}

	return user, nil
}

func NewAuthService(userRepository repositories.UserRepository) AuthService {
	return &AuthServiceImpl{
		userRepository: userRepository,
	}
}

func (s *AuthServiceImpl) ValidateOAuthToken(
	ctx context.Context,
	provider string,
	accessToken string,
) (*daos.User, error) {
	// TODO: CONFIGURE THIS
	const aud = "834953141481-an55r41f085lol5fknij3rp5g9e8ho19.apps.googleusercontent.com"

	payload, err := idtoken.Validate(ctx, accessToken, aud)
	if err != nil {
		return nil, err
	}

	claims := parseOAuthClaims(payload.Claims)

	// See if user exists
	user, err := s.userRepository.GetUserByEmail(ctx, payload.Claims["email"].(string))

	if err != nil && errors.Is(err, repositories.ErrNoRecord) {
		password, err := utils.GenerateRandomString(32)
		if err != nil {
			return nil, err
		}

		// Create user
		user, err = s.userRepository.CreateUser(
			ctx, &contracts.CreateUserRequest{
				FirstName: claims.FirstName,
				LastName:  claims.LastName,
				Email:     claims.Email,
				Password:  password,
			},
		)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

func parseOAuthClaims(claims map[string]interface{}) *OAuthClaims {
	return &OAuthClaims{
		Email:     claims["email"].(string),
		FirstName: claims["given_name"].(string),
		LastName:  claims["family_name"].(string),
	}
}
