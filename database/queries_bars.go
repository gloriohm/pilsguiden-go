package database

import (
	"context"
	"fmt"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BarsFilter struct {
	Order     int
	Breweries []string
	MaxPrice  *int
	MinPrice  *int
}

func createFilteredQuery(f BarsFilter) (string, []any) {
	query := "SELECT * FROM bars WHERE 1=1"
	args := []any{}
	i := 1

	if f.MaxPrice != nil {
		query += fmt.Sprintf(" AND current_pint <= $%d", i)
	}
	return query, args
}

func setSortOrder(f BarsFilter) string {
	switch f.Order {
	case 0:
		return " ORDER BY current_pint ASC"
	case 1:
		return " ORDER BY current_pint DEC"
	case 2:
		return " ORDER BY bar ASC"
	case 3:
		return " ORDER BY bar DEC"
	default:
		return ""
	}
}

func GetBars(ctx context.Context, pool *pgxpool.Pool, f BarsFilter) ([]models.BarView, error) {
	query := "SELECT * FROM bars WHERE 1=1"
	filter, args := createFilteredQuery(f)
	query += filter

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bars, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.BarView])

	if err != nil {
		return bars, fmt.Errorf("iterating rows: %w", err)
	}

	return bars, nil
}
