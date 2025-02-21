package app

import (
	"context"
	"fmt"
	"github.com/IT-Nick/internal/infra/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

// InitDatabase устанавливает подключение к базе данных
func InitDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	const op = "app.InitDatabase"

	connConfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse database config: %w", op, err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create database pool: %w", op, err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %w", op, err)
	}

	log.Println("Database connected successfully!")
	return db, nil
}
