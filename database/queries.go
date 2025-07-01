package database

import (
	"context"
	"fmt"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
)

// Get current bars
func getBarsByLocation(conn *pgx.Conn, column string, id int) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar

	// Be very cautious: don't format raw input into SQL without validation
	query := fmt.Sprintf(`SELECT bar, size, current_pint FROM current_bars_view WHERE %s = $1`, column)

	rows, err := conn.Query(ctx, query, id)
	if err != nil {
		return bars, err
	}
	defer rows.Close()

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

func GetBarsByFylke(conn *pgx.Conn, fylke int) ([]models.Bar, error) {
	return getBarsByLocation(conn, "fylke", fylke)
}

func GetBarsByKommune(conn *pgx.Conn, kommune int) ([]models.Bar, error) {
	return getBarsByLocation(conn, "sted", kommune)
}

func GetBarsBySted(conn *pgx.Conn, sted int) ([]models.Bar, error) {
	return getBarsByLocation(conn, "nabolag", sted)
}

// Get single bar
func GetBarBySlug(conn *pgx.Conn, slug string) (*models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT bar, size, pint, address FROM bars WHERE slug = $1`
	row := conn.QueryRow(ctx, query, slug)

	var bar models.Bar
	if err := row.Scan(&bar.Name, &bar.Size, &bar.Pint, &bar.Address); err != nil {
		return nil, fmt.Errorf("db scan: %w", err)
	}

	return &bar, nil
}

func GetAboutPageData(conn *pgx.Conn) (*models.AboutInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT COUNT(*) AS total, MIN(current_pint), MAX(current_pint) FROM current_bars_view`
	row := conn.QueryRow(ctx, query)

	var about models.AboutInfo
	if err := row.Scan(&about.Total, &about.MinPrice, &about.MaxPrice); err != nil {
		return nil, err
	}
	diff := about.MaxPrice - about.MinPrice
	about.Diff = diff
	return &about, nil
}

// Get locations by hierarchy queries
func GetFylker(conn *pgx.Conn) ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var fylker []models.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'fylke'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return fylker, err
	}

	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Slug, &location.Hierarchy, &location.Parent); err != nil {
			return fylker, fmt.Errorf("scanning row: %w", err)
		}
		fylker = append(fylker, location)
	}

	if rows.Err() != nil {
		return fylker, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return fylker, nil
}

func GetKommuner(conn *pgx.Conn) ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var kommuner []models.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'sted'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return kommuner, err
	}

	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Slug, &location.Hierarchy, &location.Parent); err != nil {
			return kommuner, fmt.Errorf("scanning row: %w", err)
		}
		kommuner = append(kommuner, location)
	}

	if rows.Err() != nil {
		return kommuner, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return kommuner, nil
}

func GetSteder(conn *pgx.Conn) ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var steder []models.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'nabolag'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return steder, err
	}

	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Slug, &location.Hierarchy, &location.Parent); err != nil {
			return steder, fmt.Errorf("scanning row: %w", err)
		}
		steder = append(steder, location)
	}

	if rows.Err() != nil {
		return steder, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return steder, nil
}

// Statistical Queries
func GetTotalBars(conn *pgx.Conn) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT COUNT(*) AS total FROM current_bars_view`
	row := conn.QueryRow(ctx, query)

	var total int
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func GetTopTen(conn *pgx.Conn) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar
	query := `SELECT bar, size, current_pint FROM current_bars_view ORDER BY current_pint ASC LIMIT 10`

	rows, err := conn.Query(ctx, query)
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

func GetBottomTen(conn *pgx.Conn) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar
	query := `SELECT bar, size, current_pint FROM current_bars_view ORDER BY current_pint DESC LIMIT 10`

	rows, err := conn.Query(ctx, query)
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

func GetBreweries(conn *pgx.Conn) ([]models.Brewery, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var breweries []models.Brewery
	query := `SELECT id, name, popular FROM breweries`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return breweries, err
	}

	for rows.Next() {
		var brewery models.Brewery
		if err := rows.Scan(&brewery.ID, &brewery.Name, &brewery.Popular); err != nil {
			return breweries, fmt.Errorf("scanning row: %w", err)
		}
		breweries = append(breweries, brewery)
	}

	if rows.Err() != nil {
		return breweries, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return breweries, nil
}
