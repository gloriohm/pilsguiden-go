package database

import (
	"context"
	"fmt"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetBarsByLocation(conn *pgx.Conn, id int, column string) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar

	query := fmt.Sprintf(`SELECT * FROM current_bars WHERE %s = $1`, column)
	rows, err := conn.Query(ctx, query, id)
	if err != nil {
		return bars, err
	}
	defer rows.Close()

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(
			&bar.ID, &bar.Name, &bar.Price, &bar.Size, &bar.Pint, &bar.PriceChecked,
			&bar.Address, &bar.Fylke, &bar.FylkeName, &bar.FylkeSlug,
			&bar.Kommune, &bar.KommuneName, &bar.KommuneSlug,
			&bar.Sted, &bar.StedName, &bar.StedSlug,
			&bar.Flyplass, &bar.Brewery, &bar.Latitude, &bar.Longitude,
			&bar.CurrentPint, &bar.CurrentPrice,
			&bar.FromTime, &bar.UntilTime, &bar.HappyChecked,
		); err != nil {
			return bars, fmt.Errorf("scanning row: %w", err)
		}
		bars = append(bars, bar)
	}

	if rows.Err() != nil {
		return bars, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return bars, nil
}

func GetBarsByLocationAndTime(conn *pgx.Conn, id int, column, date, customTime string) ([]models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var bars []models.Bar

	// Be very cautious: don't format raw input into SQL without validation
	query := getBarsByTimeQuery

	rows, err := conn.Query(ctx, query, date, customTime, id, column)
	if err != nil {
		return bars, err
	}
	defer rows.Close()

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(
			&bar.ID, &bar.Name, &bar.Price, &bar.Size, &bar.Pint, &bar.PriceChecked,
			&bar.Address, &bar.Fylke, &bar.FylkeName, &bar.FylkeSlug,
			&bar.Kommune, &bar.KommuneName, &bar.KommuneSlug,
			&bar.Sted, &bar.StedName, &bar.StedSlug,
			&bar.Flyplass, &bar.Brewery, &bar.Latitude, &bar.Longitude,
			&bar.CurrentPint, &bar.CurrentPrice,
			&bar.FromTime, &bar.UntilTime, &bar.HappyChecked,
		); err != nil {
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
	query := `SELECT COUNT(*) AS total FROM current_bars`
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
	query := `SELECT bar, size, current_pint FROM current_bars ORDER BY current_pint ASC LIMIT 10`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return bars, err
	}

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(&bar.Name, &bar.Size, &bar.CurrentPint); err != nil {
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
	query := `SELECT bar, size, current_pint FROM current_bars ORDER BY current_pint DESC LIMIT 10`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return bars, err
	}

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(&bar.Name, &bar.Size, &bar.CurrentPint); err != nil {
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

const getBarsByTimeQuery = `
WITH vars AS (
    SELECT $1::date AS current_date, $2::time AS current_time
),
hk AS (
    SELECT *
    FROM happy_keys, vars
    WHERE passes_midnight IS false
      AND (day & (1 << EXTRACT(DOW FROM vars.current_date)::INT)) > 0
      AND vars.current_time >= from_time
      AND vars.current_time < until_time
    UNION ALL
    SELECT *
    FROM happy_keys, vars
    WHERE passes_midnight IS true
      AND (
           ((day & (1 << EXTRACT(DOW FROM vars.current_date)::INT)) > 0 AND vars.current_time >= from_time)
           OR
           ((end_day & (1 << EXTRACT(DOW FROM vars.current_date)::INT)) > 0 AND vars.current_time < until_time)
      )
)
SELECT 
    b.id, b.bar, b.price, b.size, b.pint, b.price_checked,
    b.address, b.fylke, l_fylke.name AS fylke_name, l_fylke.slug AS fylke_slug,
    b.sted, l_kommune.name AS kommune_name, l_kommune.slug AS kommune_slug,
    b.nabolag, l_sted.name AS sted_name, l_sted.slug AS sted_slug,
    b.flyplass, b.brewery, b.latitude, b.longitude,
    CASE WHEN b.timed_prices AND hk.pint IS NOT NULL THEN hk.pint ELSE b.pint END AS current_pint,
    CASE WHEN b.timed_prices AND hk.price IS NOT NULL THEN hk.price ELSE b.price END AS current_price,
    hk.from_time, hk.until_time, hk.price_checked AS hk_checked
FROM bars b
LEFT JOIN hk ON b.id = hk.bar
LEFT JOIN locations l_fylke ON l_fylke.id = b.fylke
LEFT JOIN locations l_kommune ON l_kommune.id = b.sted
LEFT JOIN locations l_sted ON l_sted.id = b.nabolag
WHERE b.is_active IS true
  AND (
      CASE 
          WHEN $4 = 'fylke' THEN b.fylke 
          WHEN $4 = 'kommune' THEN b.sted 
          WHEN $4 = 'sted' THEN b.nabolag 
      END
  ) = $3
ORDER BY current_pint ASC
`
