// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"github.com/rapidaai/api/integration-api/config"
	integration_routers "github.com/rapidaai/api/integration-api/router"
	web_client "github.com/rapidaai/pkg/clients/web"
	"github.com/rapidaai/pkg/middlewares"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/rapidaai/pkg/authenticators"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
)

// wrapper for gin engine
type AppRunner struct {
	E         *gin.Engine
	S         *grpc.Server
	Cfg       *config.IntegrationConfig
	Logger    commons.Logger
	Postgres  connectors.PostgresConnector
	Redis     connectors.RedisConnector
	Closeable []func(context.Context) error
}

func main() {
	ctx := context.Background()
	appRunner := AppRunner{E: gin.New(), S: grpc.NewServer()}

	if err := appRunner.ResolveConfig(); err != nil {
		panic(err)
	}

	appRunner.Logging()
	appRunner.AllConnectors()

	if err := appRunner.Migrate(); err != nil {
		appRunner.Logger.Errorf("Warning: Migration failed: %v", err)
		panic(err)
	}

	authClient := web_client.NewAuthenticator(&appRunner.Cfg.AppConfig, appRunner.Logger, appRunner.Redis)
	appRunner.S = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewares.NewRequestLoggerUnaryServerMiddleware(appRunner.Cfg.Name, appRunner.Logger),
			middlewares.NewRecoveryUnaryServerMiddleware(appRunner.Logger),
			middlewares.NewServiceAuthenticatorUnaryServerMiddleware(
				authenticators.NewServiceAuthenticator(&appRunner.Cfg.AppConfig, appRunner.Logger, appRunner.Postgres),
				appRunner.Logger,
			),
			middlewares.NewProjectAuthenticatorUnaryServerMiddleware(
				authenticators.NewProjectAuthenticator(&appRunner.Cfg.AppConfig, appRunner.Logger, authClient),
				appRunner.Logger,
			),
		),
		grpc.ChainStreamInterceptor(
			middlewares.NewRequestLoggerStreamServerMiddleware(appRunner.Cfg.Name, appRunner.Logger),
			middlewares.NewRecoveryStreamServerMiddleware(appRunner.Logger),
			middlewares.NewServiceAuthenticatorStreamServerMiddleware(
				authenticators.NewServiceAuthenticator(&appRunner.Cfg.AppConfig, appRunner.Logger, appRunner.Postgres),
				appRunner.Logger,
			),
			middlewares.NewProjectAuthenticatorStreamServerMiddleware(
				authenticators.NewProjectAuthenticator(&appRunner.Cfg.AppConfig, appRunner.Logger, authClient),
				appRunner.Logger,
			),
		),
	)

	if err := appRunner.Init(ctx); err != nil {
		panic(err)
	}

	appRunner.AllRouters()
	appRunner.AllMiddlewares()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", appRunner.Cfg.Host, appRunner.Cfg.Port))
	if err != nil {
		log.Fatalf("Failed to create connection tcp %v", err)
	}

	defer appRunner.Close(ctx)
	cmuxListener := cmux.New(listener)

	http2GRPCFilteredListener := cmuxListener.Match(cmux.HTTP2())
	grpcFilteredListener := cmuxListener.Match(
		cmux.HTTP1HeaderField("Content-type", "application/grpc-web+proto"),
		cmux.HTTP1HeaderField("x-grpc-web", "1"),
	)
	rpcFilteredListener := cmuxListener.Match(cmux.Any())

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return appRunner.E.RunListener(rpcFilteredListener)
	})

	group.Go(func() error {
		wrappedServer := grpcweb.WrapServer(appRunner.S, grpcweb.WithOriginFunc(func(origin string) bool { return true }))
		httpServer := http.Server{
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				wrappedServer.ServeHTTP(resp, req)
			}),
		}
		return httpServer.Serve(grpcFilteredListener)
	})

	group.Go(func() error {
		return appRunner.S.Serve(http2GRPCFilteredListener)
	})

	if err := cmuxListener.Serve(); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}

// -------------------- FIXED MIGRATION LOGIC --------------------

func (app *AppRunner) Migrate() error {
	skipMigration := flag.Bool("skip-migration", false, "Skip migration")
	flag.Parse()
	if *skipMigration {
		app.Logger.Infof("Skipping migration")
		return nil
	}

	var dsn string

	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn = v
	} else if v := os.Getenv("DATABASE_DSN"); v != "" {
		dsn = v
	} else if v := os.Getenv("POSTGRES_DSN"); v != "" {
		dsn = v
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			app.Cfg.PostgresConfig.Auth.User,
			app.Cfg.PostgresConfig.Auth.Password,
			app.Cfg.PostgresConfig.Host,
			app.Cfg.PostgresConfig.Port,
			app.Cfg.PostgresConfig.DBName,
			app.Cfg.PostgresConfig.SslMode,
		)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationsPath := fmt.Sprintf("file://%s/api/integration-api/migrations", currentDir)

	app.Logger.Infof("Using DSN for migration: %s", dsn)

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	app.Logger.Infof("Migrations completed successfully")
	return nil
}

// -------------------- HELPERS --------------------

func (app *AppRunner) Logging() error {
	logger, err := commons.NewApplicationLogger(
		commons.Level(app.Cfg.LogLevel),
		commons.Name(app.Cfg.Name),
	)
	if err != nil {
		return err
	}
	app.Logger = logger
	return nil
}

func (g *AppRunner) AllConnectors() {
	g.Postgres = connectors.NewPostgresConnector(&g.Cfg.PostgresConfig, g.Logger)
	g.Redis = connectors.NewRedisConnector(&g.Cfg.RedisConfig, g.Logger)
}

func (app *AppRunner) ResolveConfig() error {
	vConfig, err := config.InitConfig()
	if err != nil {
		return err
	}
	cfg, err := config.GetApplicationConfig(vConfig)
	if err != nil {
		return err
	}
	app.Cfg = cfg
	gin.SetMode(gin.ReleaseMode)
	return nil
}

func (app *AppRunner) Init(ctx context.Context) error {
	if err := app.Postgres.Connect(ctx); err != nil {
		return err
	}
	if err := app.Redis.Connect(ctx); err != nil {
		return err
	}
	app.Closeable = append(app.Closeable, app.Redis.Disconnect, app.Postgres.Disconnect)
	return nil
}

func (app *AppRunner) Close(ctx context.Context) {
	for _, c := range app.Closeable {
		_ = c(ctx)
	}
}

func (g *AppRunner) AllRouters() {
	integration_routers.HealthCheckRoutes(g.Cfg, g.E, g.Logger, g.Postgres)
	integration_routers.ProviderApiRoute(g.Cfg, g.S, g.Logger, g.Postgres)
	integration_routers.AuditLoggingApiRoute(g.Cfg, g.S, g.Logger, g.Postgres)
}

func (g *AppRunner) AllMiddlewares() {
	g.E.Use(gin.Recovery())
	g.E.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:    []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Access-Control-Allow-Origin", "X-Grpc-Web"},
	}))
}
