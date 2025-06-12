package database

import (
	"context"
	"fmt"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
)

// Get current bars
func GetBarsByFylke(conn *pgx.Conn, fylke int) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar
	query := `SELECT bar, size, current_pint FROM current_bars_view WHERE fylke = $1`

	rows, err := conn.Query(ctx, query, fylke)
	if err != nil {
		return bars, err
	}

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(&bar.Name, &bar.Size, &bar.Pint); err != nil {
			return bars, fmt.Errorf("scanning row: %w", err)
		}
		bars = append(bars, bar)
	}

	if rows.Err() != nil {
		return bars, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return bars, nil
}

// Get single bar
func GetBarBySlug(conn *pgx.Conn, slug string) (*models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT bar, size, pint FROM bars WHERE slug = $1`

	row := conn.QueryRow(ctx, query, slug)
	var bar models.Bar
	if err := row.Scan(&bar.Name, &bar.Size, &bar.Pint); err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}

	return &bar, nil
}
