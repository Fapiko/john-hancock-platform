package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/persistence/graphdb"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/crypto/bcrypt"
)

var sessionExpiration = 24 * time.Hour

type Controller struct {
	UserRepository repositories.Repository
}

func NewController(userRepository repositories.Repository) *Controller {
	return &Controller{
		UserRepository: userRepository,
	}
}

func (c *Controller) createUserHandler(w http.ResponseWriter, r *http.Request) {
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

func (c *Controller) loginUserHandler(w http.ResponseWriter, r *http.Request) {
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

	session := &contracts.SessionResponse{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Expires:   time.Now().Add(sessionExpiration),
	}
	err = c.UserRepository.CreateSession(ctx, user.ID, session)
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

func (c *Controller) SetupRoutes(ctx context.Context, router *swagger.Router) {
	log := logger.Get(ctx)

	// Create User
	_, err := router.AddRoute(http.MethodPost, "/users", c.createUserHandler, swagger.Definitions{
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
	})
	if err != nil {
		log.WithError(err).Error("Error creating route")
	}

	_, err = router.AddRoute(http.MethodPost, "/users/auth", c.loginUserHandler, swagger.Definitions{
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
	})
	if err != nil {
		log.WithError(err).Error("Error creating route")
	}
}
