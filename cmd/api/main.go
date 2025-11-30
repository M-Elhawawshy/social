package main

import (
	"context"
	"social/internal/env"
	"social/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

const version = "0.0.1"

func main() {
	addr := env.GetString("ADDR", ":8080")
	cfg := config{
		addr: addr,
		env:  env.GetString("ENV", "development"),
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
	app := &application{
		config: cfg,
		models: dbModels,
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
