package main

import (
	"context"
	"net"
	"net/http"
	"time"

	stdLog "log"

	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/users"
	"github.com/gorilla/handlers"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/sirupsen/logrus"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

func main() {
	log := logrus.New()
	ctx := logger.WithLogger(context.Background(), log)
	muxRouter := mux.NewRouter()

	muxRouter.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(http.Dir("swaggerui"))))

	router, err := swagger.NewRouter(apirouter.NewGorillaMuxRouter(muxRouter), swagger.Options{
		Context: ctx,
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   "John Hancock",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Error creating router")
	}

	neo4jDriver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "pwd123", ""))
	if err != nil {
		log.WithError(err).Error("Error creating neo4j driver")
	}

	neo4jSession := neo4jDriver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		_ = neo4jSession.Close()
	}()

	userRepository := users.NewRepositoryNeo4j(neo4jSession)
	userController := users.NewController(userRepository)
	userController.SetupRoutes(ctx, router)

	err = router.GenerateAndExposeSwagger()
	if err != nil {
		log.WithError(err).Error("Error generating swagger")
	}

	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Language", "Content-Type", "Origin"}),
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}))(muxRouter)

	srv := &http.Server{
		Addr:         ":11000",
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		ErrorLog:     stdLog.New(log.Writer(), "", 0),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	log.Infof("Swagger up and running at http://0.0.0.0%s/swagger/", srv.Addr)
	err = srv.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("Error starting server")
	}
}
