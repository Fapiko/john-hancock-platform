package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	stdLog "log"

	"github.com/fapiko/john-hancock-platform/app/config"
	"github.com/fapiko/john-hancock-platform/app/context/logger"
	"github.com/fapiko/john-hancock-platform/app/controllers"
	"github.com/fapiko/john-hancock-platform/app/repositories"
	"github.com/fapiko/john-hancock-platform/app/services"
	"github.com/fapiko/john-hancock-platform/app/users"
	"github.com/gorilla/handlers"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

func main() {
	log := logrus.New()
	ctx := logger.WithLogger(context.Background(), log)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Panic(err)
	}

	muxRouter := mux.NewRouter()

	muxRouter.PathPrefix("/swagger/").Handler(
		http.StripPrefix(
			"/swagger/",
			http.FileServer(http.Dir("swaggerui")),
		),
	)

	router, err := swagger.NewRouter(
		apirouter.NewGorillaMuxRouter(muxRouter), swagger.Options{
			Context: ctx,
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "John Hancock",
					Version: "1.0.0",
				},
				Components: openapi3.Components{
					SecuritySchemes: openapi3.SecuritySchemes{
						"apiKey": &openapi3.SecuritySchemeRef{
							Value: &openapi3.SecurityScheme{
								Type: "apiKey",
								In:   "header",
								Name: "Authorization",
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.WithError(err).Error("Error creating router")
	}

	var userRepository repositories.UserRepository
	if cfg.Database.Type == config.DB_TYPE_NEO4J {
		neo4jDriver, err := neo4j.NewDriver(
			"bolt://localhost:7687",
			neo4j.BasicAuth("neo4j", "pwd123", ""),
		)
		if err != nil {
			log.WithError(err).Error("Error creating neo4j driver")
		}
		defer func() {
			err := neo4jDriver.Close()
			if err != nil {
				log.WithError(err).Error("Error closing neo4j driver")
			}
		}()

		userRepository = repositories.NewUserRepositoryNeo4j(neo4jDriver)
	} else {
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Hostname,
			cfg.Database.Name,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		userRepository = repositories.NewUserRepositoryMySql(db)
	}

	authService := services.NewAuthService(userRepository)

	userController := controllers.NewController(userRepository, authService)
	userController.SetupRoutes(ctx, router)

	sessionWorker := users.NewSessionWorker(userRepository)
	go sessionWorker.Start(ctx)

	err = router.GenerateAndExposeSwagger()
	if err != nil {
		log.WithError(err).Error("Error generating swagger")
	}

	corsHandler := handlers.CORS(
		handlers.AllowedHeaders(
			[]string{
				"Accept",
				"Accept-Language",
				"Authorization",
				"Content-Language",
				"Content-Type",
				"Origin",
			},
		),
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
	)(muxRouter)

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
