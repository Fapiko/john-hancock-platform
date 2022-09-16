package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/persistence/graphdb"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/fapiko/john-hancock-platform/app/services"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	UserRepository repositories.UserRepository
	AuthService    services.AuthService
}

func NewController(
	userRepository repositories.UserRepository,
	authService services.AuthService,
) *UserController {
	return &UserController{
		UserRepository: userRepository,
		AuthService:    authService,
	}
}

func (c *UserController) createUserHandler(w http.ResponseWriter, r *http.Request) {
	createUserReq := &contracts.CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if createUserReq.Password == "" ||
		createUserReq.FirstName == "" ||
		createUserReq.LastName == "" ||
		createUserReq.Email == "" {

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log := logger.Get(r.Context())
	user, err := c.UserRepository.CreateUser(r.Context(), createUserReq)
	if err != nil {
		if err.(*neo4j.Neo4jError).Code == graphdb.SchemaFailedCode {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Error("failed to create user")
		}
		return
	}

	resp, err := json.Marshal(user.ToResponse())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("failed to marshal user")
		return
	}

	_, _ = w.Write(resp)
}

func (c *UserController) getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := c.UserRepository.GetUserBySessionID(r.Context(), sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(user.ToResponse())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(resp)
}

func (c *UserController) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)
	loginRequest := &contracts.LoginUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := c.UserRepository.GetUserByEmail(ctx, loginRequest.Email)
	if err != nil {
		log.WithError(err).Error("failed to get user")

		// Throw unauthorized to avoid leaking user existence
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		// Only log if we had a bcrypt failure
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.WithError(err).Error("failed to compare password")
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := c.UserRepository.CreateSession(ctx, user.ID)
	if err != nil {
		log.WithError(err).Error("failed to create session")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respObject := &contracts.LoginUserResponse{
		Session: session,
		User:    user.ToResponse(),
	}

	resp, err := json.Marshal(respObject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("failed to marshal login response")
		return
	}

	_, _ = w.Write(resp)
}

func (c *UserController) validateOauth2Token(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()

	validationRequest := &contracts.OAuthValidateRequest{}
	if err := json.NewDecoder(r.Body).Decode(validationRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := c.AuthService.ValidateOAuthToken(
		r.Context(),
		validationRequest.Provider,
		validationRequest.AccessToken,
	)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := c.UserRepository.CreateSession(r.Context(), user.ID)

	respObject := &contracts.LoginUserResponse{
		Session: session,
		User:    user.ToResponse(),
	}

	resp, err := json.Marshal(respObject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("failed to marshal login response")
		return
	}

	_, _ = w.Write(resp)
}

func (c *UserController) SetupRoutes(ctx context.Context, router *swagger.Router) {
	log := logger.Get(ctx)

	// Get current user
	var err error

	operation := swagger.NewOperation()
	operation.Security = &openapi3.SecurityRequirements{
		{
			"apiKey": {},
		},
	}
	operation.AddResponse(
		http.StatusOK, &openapi3.Response{},
	)

	_, err = router.AddRawRoute(
		http.MethodGet,
		"/user",
		c.getCurrentUserHandler,
		operation,
	)

	// Create User
	_, err = router.AddRoute(
		http.MethodPost, "/users", c.createUserHandler, swagger.Definitions{
			RequestBody: &swagger.ContentValue{
				Content: swagger.Content{
					"application/json": {Value: contracts.CreateUserRequest{}},
				},
				Description: "Create a new user",
			},
			Responses: map[int]swagger.ContentValue{
				http.StatusOK: {
					Content: swagger.Content{
						"application/json": {Value: contracts.UserResponse{}},
					},
					Description: "User created",
				},
				http.StatusInternalServerError: {
					Description: "Internal server error",
				},
				http.StatusConflict: {
					Description: "User already exists",
				},
				http.StatusBadRequest: {
					Description: "Bad Request",
				},
			},
		},
	)
	if err != nil {
		log.WithError(err).Error("Error creating route")
	}

	_, err = router.AddRoute(
		http.MethodPost, "/users/auth", c.loginUserHandler, swagger.Definitions{
			RequestBody: &swagger.ContentValue{
				Content: swagger.Content{
					"application/json": {Value: contracts.LoginUserRequest{}},
				},
				Description: "Authenticates a user",
			},
			Responses: map[int]swagger.ContentValue{
				http.StatusOK: {
					Content: swagger.Content{
						"application/json": {Value: contracts.UserResponse{}},
					},
					Description: "User authenticated",
				},
				http.StatusInternalServerError: {
					Description: "Internal server error",
				},
				http.StatusBadRequest: {
					Description: "Bad Request",
				},
			},
		},
	)

	_, err = router.AddRoute(
		http.MethodPost, "/oauth2/token", c.validateOauth2Token, swagger.Definitions{
			RequestBody: &swagger.ContentValue{
				Content: swagger.Content{
					"application/json": {Value: contracts.OAuthValidateRequest{}},
				},
				Description: "Validates an oauth2 token",
			},
			Responses: map[int]swagger.ContentValue{
				http.StatusOK: {
					Content: swagger.Content{
						"application/json": {},
					},
					Description: "Successful validation",
				},
			},
		},
	)

	if err != nil {
		log.WithError(err).Error("Error creating route")
	}
}
