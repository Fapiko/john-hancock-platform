package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/support/gorilla"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/services"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type KeyController struct {
	authService services.AuthService
	keyService  services.KeyService
}

func NewKeyController(
	authService services.AuthService,
	keyService services.KeyService,
) *KeyController {
	return &KeyController{
		authService: authService,
		keyService:  keyService,
	}
}

func (c *KeyController) getKeyTypesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	err := json.NewEncoder(w).Encode(contracts.KeyAlgorithmStrings)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("failed to encode key types")
		return
	}
}

func (c *KeyController) createKeyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := &contracts.CreateKeyRequest{}
	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.WithError(err).Error("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := c.keyService.CreateKey(ctx, user.ID, req.Name, req.Algorithm, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *KeyController) getKeysHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := c.keyService.GetKeysForUser(ctx, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *KeyController) RegisterRoutes(
	ctx context.Context,
	router *swagger.Router[gorilla.HandlerFunc, *mux.Route],
) {
	var log = logger.Get(ctx)

	securityRequirements := swagger.SecurityRequirements{
		{
			"apiKey": {},
		},
	}

	var err error

	_, err = router.AddRoute(http.MethodPut, "/keys", c.createKeyHandler, swagger.Definitions{})
	if err != nil {
		log.WithError(err).Fatal("failed to add route")
	}

	_, err = router.AddRoute(
		http.MethodGet,
		"/keys/types",
		c.getKeyTypesHandler,
		swagger.Definitions{Security: securityRequirements},
	)
	if err != nil {
		log.WithError(err).Fatal("failed to add route")
	}

	_, err = router.AddRoute(
		http.MethodGet,
		"/keys",
		c.getKeysHandler,
		swagger.Definitions{Security: securityRequirements},
	)
	if err != nil {
		log.WithError(err).Fatal("failed to add route")
	}
}
