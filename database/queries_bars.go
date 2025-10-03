package database

import (
	"context"
	"fmt"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createFilteredQuery(f models.BarsFilter) (string, []any) {
	query := ""
	args := []any{}
	i := 1

	if f.MaxPrice != nil {
		query += fmt.Sprintf(" AND current_pint <= $%d", i)
		i++
	}

	if f.MaxPrice != nil {
		query += fmt.Sprintf(" AND current_pint >= $%d", i)
		i++
	}

	if len(f.Breweries) >= 1 {
		query += fmt.Sprintf(" AND brewery = ANY($%d)", i)
		i++
	}

	query += setSortOrder(f)

	return query, args
}

func setSortOrder(f models.BarsFilter) string {
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

func GetBars(ctx context.Context, pool *pgxpool.Pool, f models.BarsFilter) ([]models.BarView, error) {
	query := "SELECT * FROM bars WHERE 1=1"
	filter, args := createFilteredQuery(f)
	query += filter

	fmt.Println(query)

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
