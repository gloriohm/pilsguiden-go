package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-router/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

type app struct {
	Pool *pgxpool.Pool
}

var sessionStore = cache.New(30*time.Minute, 10*time.Minute)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PW"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	ctx := context.Background()
	pool, err := NewDBPool(ctx, dsn)

	if err != nil {
		log.Fatal(err)
	}

	app := &app{Pool: pool}
	database.InitStaticData(app.Pool)

	srv := http.Server{
		Addr:    ":3000",
		Handler: app.routes(),
	}

	log.Println("Listening on :3000")
	srv.ListenAndServe()
}

func NewDBPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	cfg.MinConns = 1
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
