// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package connectors

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-gorm/caches/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	commons "github.com/rapidaai/pkg/commons"
	configs "github.com/rapidaai/pkg/configs"
)

type PostgresConnector interface {
	Connector
	Query(ctx context.Context, qry string, dest interface{}) error
	DB(ctx context.Context) *gorm.DB
}

type postgresConnector struct {
	logger commons.Logger
	cfg    *configs.PostgresConfig
	db     *gorm.DB
}

func NewPostgresConnector(config *configs.PostgresConfig, logger commons.Logger) PostgresConnector {
	return &postgresConnector{cfg: config, logger: logger}
}

func (psql *postgresConnector) DB(ctx context.Context) *gorm.DB {
	return psql.db.WithContext(ctx)
}

/*
	resolveConnectionString priority order:

	1. DATABASE_URL / DATABASE_DSN / POSTGRES_DSN
	2. POSTGRES_* / PG* env vars
	3. Config file (existing behavior)
*/
func (psql *postgresConnector) connectionString() string {
	// 1️⃣ Full DSN envs (Railway default)
	if v := firstNonEmpty(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_DSN"),
		os.Getenv("POSTGRES_DSN"),
	); v != "" {
		psql.logger.Infof("Using Postgres DSN from environment")
		return v
	}

	// 2️⃣ Build from POSTGRES_* or PG*
	host := firstNonEmpty(os.Getenv("POSTGRES_HOST"), os.Getenv("PGHOST"))
	port := firstNonEmpty(os.Getenv("POSTGRES_PORT"), os.Getenv("PGPORT"))
	user := firstNonEmpty(os.Getenv("POSTGRES_USER"), os.Getenv("PGUSER"))
	password := firstNonEmpty(os.Getenv("POSTGRES_PASSWORD"), os.Getenv("PGPASSWORD"))
	dbname := firstNonEmpty(os.Getenv("POSTGRES_DB"), os.Getenv("PGDATABASE"))
	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	if host != "" && user != "" && dbname != "" {
		if port == "" {
			port = "5432"
		}
		psql.logger.Infof("Using Postgres connection from env vars (POSTGRES_*/PG*)")
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbname, port, sslmode,
		)
	}

	// 3️⃣ Fallback to config (existing behavior)
	psql.logger.Warnf("Falling back to Postgres config file values")
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		psql.cfg.Host,
		psql.cfg.Auth.User,
		psql.cfg.Auth.Password,
		psql.cfg.DBName,
		psql.cfg.Port,
		psql.cfg.SslMode,
	)
}

func (psql *postgresConnector) Connect(ctx context.Context) error {
	lgr := logger.Discard.LogMode(logger.Silent)

	dsn := psql.connectionString()
	psql.logger.Infof("Connecting to Postgres")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: lgr,
	})
	if err != nil {
		psql.logger.Errorf("Failed to open postgres connection: %v", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		psql.logger.Errorf("Failed to create postgres client connection pool: %v", err)
		return err
	}

	sqlDB.SetMaxIdleConns(psql.cfg.MaxIdealConnection)
	sqlDB.SetMaxOpenConns(psql.cfg.MaxOpenConnection)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Second-level cache (unchanged)
	if psql.cfg.SLCache != nil {
		psql.logger.Debugf("Second level caching is enabled for gorm")
		rdb := NewRedisPostgresCacheConnector(psql.cfg.SLCache, psql.logger)
		if err = rdb.Connect(ctx); err == nil {
			cachesPlugin := &caches.Caches{Conf: &caches.Config{Cacher: rdb}}
			_ = db.Use(cachesPlugin)
		} else {
			psql.logger.Errorf("Unable to initialize cache connector")
		}
	}

	psql.db = db
	psql.logger.Infof("Postgres connected successfully")
	return nil
}

func (psql *postgresConnector) Name() string {
	return "Postgres"
}

func (psql *postgresConnector) IsConnected(ctx context.Context) bool {
	if psql.db == nil {
		return false
	}
	db, err := psql.db.DB()
	if err != nil {
		return false
	}
	return db.PingContext(ctx) == nil
}

func (psql *postgresConnector) Disconnect(ctx context.Context) error {
	psql.logger.Debug("Disconnecting postgres")
	db, err := psql.db.DB()
	psql.db = nil
	if err != nil {
		return err
	}
	return db.Close()
}

func (psql *postgresConnector) Query(ctx context.Context, qry string, dest interface{}) error {
	return psql.db.Raw(qry).Scan(dest).Error
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
