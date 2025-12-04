package main

import (
	"context"
	"log"
	"social/internal/env"
	"social/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	_ = godotenv.Load()
}

const version = "0.0.1"

//	@title			Social API
//	@version		1.0
//	@description	This is a Social API server
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	addr := env.GetString("ADDR", ":8080")
	cfg := config{
		addr:   addr,
		env:    env.GetString("ENV", "development"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: DBConfig{
			dsn:          env.GetString("DATABASE_URL", ""),
			maxOpenConns: 30,
		},
	}
	dbPool, err := openDB(cfg.db)
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	dbModels := models.NewModels(dbPool)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	app := &application{
		config: cfg,
		models: dbModels,
		logger: logger.Sugar(),
	}

	err = app.run()
	if err != nil {
		panic(err)
	}
}

func openDB(config DBConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(config.dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = int32(config.maxOpenConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}
