package users

import (
	"context"
	"encoding/json"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/persistence/graphdb"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Controller struct {
	UserRepository Repository
}

func NewController(userRepository Repository) *Controller {
	return &Controller{
		UserRepository: userRepository,
	}
}

func (c *Controller) createUserHandler(w http.ResponseWriter, r *http.Request) {
	createUserReq := &CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
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

func (c *Controller) SetupRoutes(ctx context.Context, router *swagger.Router) {
	_, err := router.AddRoute(http.MethodPost, "/users", c.createUserHandler, swagger.Definitions{
		RequestBody: &swagger.ContentValue{
			Content: swagger.Content{
				"application/json": {Value: CreateUserRequest{}},
			},
			Description: "Create a new user",
		},
		Responses: map[int]swagger.ContentValue{
			http.StatusOK: {
				Content: swagger.Content{
					"application/json": {Value: UserResponse{}},
				},
				Description: "User created",
			},
			http.StatusInternalServerError: {
				Description: "Internal server error",
			},
			http.StatusConflict: {
				Description: "User already exists",
			},
		},
	})
	if err != nil {
		logger.Get(ctx).WithError(err).Error("Error creating route")
	}
}
