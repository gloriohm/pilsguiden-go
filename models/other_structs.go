package models

import "github.com/jackc/pgx/v5"

type App struct {
	DB *pgx.Conn
}
