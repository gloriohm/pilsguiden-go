package app

import "github.com/jackc/pgx/v5/pgxpool"

type PostgresRepo struct{ pool *pgxpool.Pool }

func NewPostgresRepo(p *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{pool: p} }
