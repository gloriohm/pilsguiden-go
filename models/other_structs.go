package models

import "github.com/jackc/pgx/v4/pgxpool"

type App struct {
	DB *pgxpool.Pool
}
