package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/fapiko/john-hancock-platform/app/services"
)

type CertificateAuthorityController struct {
	authService           services.AuthService
	certificateService    services.CertificateService
	certificateRepository repositories.CertRepository
}

func NewCertificateAuthorityController(
	authService services.AuthService,
	certService services.CertificateService,
	certRepo repositories.CertRepository,
) *CertificateAuthorityController {
	return &CertificateAuthorityController{
		authService:           authService,
		certificateService:    certService,
		certificateRepository: certRepo,
	}
}

func (c *CertificateAuthorityController) createCAHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log := logger.Get(ctx)

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := &contracts.CreateCARequest{}
	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.WithError(err).Error("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	certData, err := c.certificateService.GenerateCert(ctx, req, services.CertTypeRootCA)
	if err != nil {
		log.WithError(err).Error("failed to generate certificate")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cert, err := c.certificateRepository.CreateCert(
		ctx,
		user.ID,
		req.Name,
		certData,
		services.CertTypeRootCA.String(),
	)

	resp := &contracts.CreateCAResponse{
		ID:      cert.ID,
		Created: cert.Created,
		Name:    cert.Name,
		Type:    cert.Type,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
	}
}

func (c *CertificateAuthorityController) SetupRoutes(ctx context.Context, router *swagger.Router) {
	log := logger.Get(ctx)

	// Get current user
	var err error

	_, err = router.AddRoute(
		http.MethodPut,
		"/certificate-authorities",
		c.createCAHandler,
		swagger.Definitions{},
	)

	if err != nil {
		log.WithError(err).Error("failed to setup routes")
	}
}
