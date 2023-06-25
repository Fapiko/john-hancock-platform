package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/support/gorilla"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/fapiko/john-hancock-platform/app/services"
	"github.com/gorilla/mux"
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

func (c *CertificateAuthorityController) getCAHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	cert, err := c.certificateService.GetCert(ctx, id)
	if err != nil {
		log.WithError(err).Error("failed to get cert")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if cert.OwnerID != user.ID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = json.NewEncoder(w).Encode(cert)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
	}
}

func (c *CertificateAuthorityController) getCAsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	certs, err := c.certificateService.GetUserCerts(
		ctx,
		user.ID,
		[]services.CertificateType{services.CertTypeRootCA, services.CertTypeIntermediateCA},
	)
	if err != nil {
		log.WithError(err).Error("failed to get certs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(certs)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
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

	certType := services.CertTypeIntermediateCA
	if req.ParentCA == "" {
		certType = services.CertTypeRootCA
	}

	certData, err := c.certificateService.CreateCACert(ctx, req, user.ID, certType)
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
		certType.String(),
		req.ParentCA,
		req.KeyID,
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

func (c *CertificateAuthorityController) createCertHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Get(ctx)

	vars := mux.Vars(r)
	certAuthorityId := vars["caId"]

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := &contracts.CreateCertificateRequest{}
	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.WithError(err).Error("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := c.certificateService.CreateCert(ctx, certAuthorityId, req, user.ID)
	if err != nil {
		log.WithError(err).Error("failed to generate certificate")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
	}
}

func (c *CertificateAuthorityController) getCertificatesHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := logger.Get(ctx)

	vars := mux.Vars(r)
	certAuthorityId := vars["caId"]

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	certs, err := c.certificateService.GetCertsByParentCAForUser(ctx, certAuthorityId, user.ID)
	if err != nil {
		log.WithError(err).Error("failed to get certs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(certs)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
	}
}

func (c *CertificateAuthorityController) getCertificateHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := logger.Get(ctx)

	vars := mux.Vars(r)
	//certAuthorityId := vars["caId"]
	certId := vars["id"]

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cert, err := c.certificateService.GetCert(ctx, certId)
	if err != nil {
		log.WithError(err).Error("failed to get cert")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if cert.OwnerID != user.ID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = json.NewEncoder(w).Encode(cert)
	if err != nil {
		log.WithError(err).Error("failed to encode response")
		return
	}
}

func (c *CertificateAuthorityController) deleteCertificateHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := logger.Get(ctx)

	vars := mux.Vars(r)
	certId := vars["id"]

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = c.certificateService.DeleteCertForUser(ctx, certId, user.ID)
	if err != nil {
		log.WithError(err).Error("failed to delete cert")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *CertificateAuthorityController) downloadCertificateHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	log := logger.Get(ctx)

	vars := mux.Vars(r)
	certId := vars["id"]

	user, err := c.authService.GetUserForRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cert, err := c.certificateService.GetCert(ctx, certId)

	certPem, err := c.certificateService.GetCertAsPEMForUser(ctx, certId, user.ID)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+cert.Name+".pem")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	_, err = w.Write([]byte(certPem))
	if err != nil {
		log.WithError(err).Error("failed to write cert")
		return
	}
}

func (c *CertificateAuthorityController) SetupRoutes(
	ctx context.Context,
	router *swagger.Router[gorilla.HandlerFunc, *mux.Route],
) {
	log := logger.Get(ctx)

	securityRequirements := swagger.SecurityRequirements{
		{
			"apiKey": {},
		},
	}

	var err error

	_, err = router.AddRoute(
		http.MethodPut,
		"/certificate-authorities",
		c.createCAHandler,
		swagger.Definitions{},
	)
	if err != nil {
		log.WithError(err).Error("failed to setup route")
	}

	_, err = router.AddRoute(
		http.MethodGet,
		"/certificate-authorities",
		c.getCAsHandler,
		swagger.Definitions{
			Querystring: swagger.ParameterValue{
				"type": swagger.Parameter{
					Description: "CA Type",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodGet,
		"/certificate-authorities/{id}",
		c.getCAHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"id": swagger.Parameter{
					Description: "CA ID",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodPut,
		"/certificate-authorities/{caId}",
		c.createCertHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"caId": swagger.Parameter{
					Description: "Certificate Authority ID",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodGet,
		"/certificate-authorities/{caId}/certificates",
		c.getCertificatesHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"caId": swagger.Parameter{
					Description: "Certificate Authority ID",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodGet,
		"/certificates/{id}",
		c.getCertificateHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"id": swagger.Parameter{
					Description: "Certificate ID",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodDelete,
		"/certificates/{id}",
		c.deleteCertificateHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"id": swagger.Parameter{
					Description: "Certificate ID",
				},
			},
			Security: securityRequirements,
		},
	)

	_, err = router.AddRoute(
		http.MethodGet,
		"/certificates/{id}/download",
		c.downloadCertificateHandler,
		swagger.Definitions{
			PathParams: swagger.ParameterValue{
				"id": swagger.Parameter{
					Description: "Certificate ID",
				},
			},
			Querystring: swagger.ParameterValue{
				"format": swagger.Parameter{
					Description: "Download format type",
				},
			},
			Security: securityRequirements,
		},
	)

	if err != nil {
		log.WithError(err).Error("failed to setup routes")
	}
}
