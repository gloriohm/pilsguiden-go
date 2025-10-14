package database

import (
	"context"
	"fmt"
	"go-router/internal/bars"
	"go-router/internal/utils"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetBarsByLocation(conn *pgxpool.Pool, id int, column string) ([]models.BarView, error) {
	valid := utils.CheckValidLocationLevel(column)
	if !valid {
		return nil, fmt.Errorf("unsupported column %q", column)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`SELECT * FROM current_bars WHERE %s = $1`, column)
	rows, err := conn.Query(ctx, query, id)
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

func GetBarsByLocationAndTime(conn *pgxpool.Pool, id int, column, date, customTime string) ([]models.BarView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := getBarsByTimeQuery

	rows, err := conn.Query(ctx, query, date, customTime, id, column)
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

// Get single bar
func GetBarBySlug(conn *pgxpool.Pool, slug string) (*models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
	SELECT 
		b.id, b.bar, b.price, b.size, b.pint, b.price_checked,
		b.address, b.fylke, l_fylke.name AS fylke_name, l_fylke.slug AS fylke_slug,
		b.sted, l_kommune.name AS kommune_name, l_kommune.slug AS kommune_slug,
		b.nabolag, l_sted.name AS sted_name, l_sted.slug AS sted_slug,
		b.flyplass, b.brewery, b.latitude, b.longitude, b.timed_prices
	FROM bars b
	LEFT JOIN locations l_fylke ON l_fylke.id = b.fylke
	LEFT JOIN locations l_kommune ON l_kommune.id = b.sted
	LEFT JOIN locations l_sted ON l_sted.id = b.nabolag
	WHERE b.is_active IS true
	AND b.slug = $1
	LIMIT 1
	`
	row := conn.QueryRow(ctx, query, slug)

	var bar models.Bar
	if err := row.Scan(&bar.ID, &bar.Name, &bar.Price, &bar.Size, &bar.Pint,
		&bar.PriceChecked, &bar.Address, &bar.Fylke, &bar.FylkeName, &bar.FylkeSlug,
		&bar.Kommune, &bar.KommuneName, &bar.KommuneSlug, &bar.Sted, &bar.StedName,
		&bar.StedSlug, &bar.Flyplass, &bar.Brewery, &bar.Latitude, &bar.Longitude, &bar.TimedPrices); err != nil {
		return nil, fmt.Errorf("db scan: %w", err)
	}

	return &bar, nil
}

func GetAboutPageData(conn *pgxpool.Pool) (*models.AboutInfo, error) {
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
func GetFylker(conn *pgxpool.Pool) ([]bars.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var fylker []bars.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'fylke'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return fylker, err
	}

	for rows.Next() {
		var location bars.Location
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

func GetKommuner(conn *pgxpool.Pool) ([]bars.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var kommuner []bars.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'sted'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return kommuner, err
	}

	for rows.Next() {
		var location bars.Location
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

func GetSteder(conn *pgxpool.Pool) ([]bars.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var steder []bars.Location
	query := `SELECT id, name, slug, hierarchy, parent FROM locs WHERE hierarchy = 'nabolag'`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return steder, err
	}

	for rows.Next() {
		var location bars.Location
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
func GetTotalBars(ctx context.Context, conn *pgxpool.Pool) (int, error) {
	query := `SELECT COUNT(*) AS total FROM current_bars`
	row := conn.QueryRow(ctx, query)

	var total int
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func GetTopTen(ctx context.Context, conn *pgxpool.Pool) ([]models.BarView, error) {
	query := `SELECT * FROM current_bars ORDER BY current_pint ASC LIMIT 10`
	rows, err := conn.Query(ctx, query)
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

func GetBottomTen(ctx context.Context, conn *pgxpool.Pool) ([]models.BarView, error) {
	query := `SELECT * FROM current_bars ORDER BY current_pint DESC LIMIT 10`
	rows, err := conn.Query(ctx, query)
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

func GetBreweries(conn *pgxpool.Pool) ([]models.Brewery, error) {
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

func GetBarMetadata(conn *pgxpool.Pool, barID int) (models.BarMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT bar_id, last_osm_sync, cuisine, opening_hours, wheelchair, website, email, phone, facebook, instagram FROM bar_metadata WHERE bar_id = $1`

	var meta models.BarMetadata
	err := conn.QueryRow(ctx, query, barID).Scan(&meta.BarID, &meta.LastOSMSync, &meta.Cuisine, &meta.OpeningHours, &meta.Wheelchair, &meta.Email, &meta.Phone, &meta.Facebook, &meta.Instagram)

	if err != nil {
		return meta, err
	}
	return meta, nil
}

func GetAllLocations(conn *pgxpool.Pool) ([]bars.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var locs []bars.Location

	query := `SELECT id, name, slug, hierarchy, parent FROM locs ORDER BY name`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return locs, err
	}

	for rows.Next() {
		var loc bars.Location
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Slug, &loc.Hierarchy, &loc.Parent); err != nil {
			return locs, fmt.Errorf("scanning row: %w", err)
		}
		locs = append(locs, loc)
	}

	if rows.Err() != nil {
		return locs, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return locs, nil
}

func GetHappyKeysByBarID(conn *pgxpool.Pool, barID int) ([]models.HappyKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var hkeys []models.HappyKey
	query := `SELECT id, bar, price, size, pint, from_time, until_time, day, updated_at, price_checked, passes_midnight, end_day FROM happy_keys WHERE bar = $1 ORDER BY day`

	rows, err := conn.Query(ctx, query, barID)
	if err != nil {
		return hkeys, err
	}

	for rows.Next() {
		var hkey models.HappyKey
		if err := rows.Scan(&hkey.ID, &hkey.BarID, &hkey.Price, &hkey.Size, &hkey.Pint, &hkey.FromTime, &hkey.UntilTime, &hkey.Day, &hkey.PriceUpdated, &hkey.PriceChecked, &hkey.PassesMidnight, &hkey.EndDay); err != nil {
			return hkeys, fmt.Errorf("scanning row: %w", err)
		}
		hkeys = append(hkeys, hkey)
	}

	if rows.Err() != nil {
		return hkeys, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return hkeys, nil
}

func CreateBrewery(conn *pgxpool.Pool, newBrew string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO breweries (name, popular) VALUES ($1, false) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, newBrew).Scan(&id)

	return err
}

func GetSearchResult(conn *pgxpool.Pool, keyword string) ([]models.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var results []models.SearchResult
	query := `SELECT id, name, slug, type
		FROM (
		SELECT
			id,
			name,
			slug,
			'/liste' AS type,
			position(LOWER($1) IN LOWER(name)) AS rank,
			LENGTH(name) AS len
		FROM locs
		WHERE name ILIKE '%' || $1 || '%'

		UNION ALL

		SELECT
			id,
			bar AS name,
			slug,
			'/bar/' AS type,
			position(LOWER($1) IN LOWER(bar)) AS rank,
			LENGTH(bar) AS len
		FROM bars
		WHERE bar ILIKE '%' || $1 || '%'
		) AS results
		ORDER BY rank, len, name
		LIMIT 20;`

	rows, err := conn.Query(ctx, query, keyword)
	if err != nil {
		return results, err
	}

	for rows.Next() {
		var r models.SearchResult
		if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Type); err != nil {
			return results, fmt.Errorf("scanning row: %w", err)
		}
		results = append(results, r)
	}

	if rows.Err() != nil {
		return results, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return results, nil
}

func UpdateBreweryWhereUnknown(conn *pgxpool.Pool, bar int, brew string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `UPDATE bars SET brewery = $1 WHERE id = $2 AND (brewery IS NULL OR brewery = 'Ukjent');`
	cmdTag, err := conn.Exec(ctx, query, brew, bar)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no update performed: brewery was not 'Ukjent' or NULL")
	}

	return nil
}

func GetBarSearchResult(conn *pgxpool.Pool, keyword string) ([]models.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var results []models.SearchResult
	query := `SELECT id, bar AS name, slug, 'bar' AS type
		FROM bars
		WHERE bar ILIKE '%' || $1 || '%'
		ORDER BY position(LOWER($1) IN LOWER(bar)), LENGTH(bar), bar
		LIMIT 20;`

	rows, err := conn.Query(ctx, query, keyword)
	if err != nil {
		return results, err
	}

	for rows.Next() {
		var r models.SearchResult
		if err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Type); err != nil {
			return results, fmt.Errorf("scanning row: %w", err)
		}
		results = append(results, r)
	}

	if rows.Err() != nil {
		return results, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return results, nil
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
    b.id, b.bar, b.price, b.size, b.pint, b.price_checked, b.slug,
    b.address, b.fylke, l_fylke.name AS fylke_name, l_fylke.slug AS fylke_slug,
    b.sted AS kommune, l_kommune.name AS kommune_name, l_kommune.slug AS kommune_slug,
    b.nabolag AS sted, l_sted.name AS sted_name, l_sted.slug AS sted_slug,
    b.flyplass, b.brewery, b.latitude, b.longitude,
    CASE WHEN b.timed_prices AND hk.pint IS NOT NULL THEN hk.pint ELSE b.pint END AS current_pint,
    CASE WHEN b.timed_prices AND hk.price IS NOT NULL THEN hk.price ELSE b.price END AS current_price,
    hk.from_time, hk.until_time, hk.price_checked AS hk_checked, hk.id AS hkey_id
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

func GetBarByID(conn *pgxpool.Pool, id int) (models.Bar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
	SELECT 
		b.id, b.bar, b.price, b.size, b.pint, b.price_checked, b.slug,
		b.address, b.fylke, l_fylke.name AS fylke_name, l_fylke.slug AS fylke_slug,
		b.sted, l_kommune.name AS kommune_name, l_kommune.slug AS kommune_slug,
		b.nabolag, l_sted.name AS sted_name, l_sted.slug AS sted_slug,
		b.flyplass, b.brewery, b.latitude, b.longitude, b.timed_prices, b.orgnummer
	FROM bars b
	LEFT JOIN locations l_fylke ON l_fylke.id = b.fylke
	LEFT JOIN locations l_kommune ON l_kommune.id = b.sted
	LEFT JOIN locations l_sted ON l_sted.id = b.nabolag
	WHERE b.id = $1
	LIMIT 1;
	`
	row := conn.QueryRow(ctx, query, id)

	var bar models.Bar
	if err := row.Scan(&bar.ID, &bar.Name, &bar.Price, &bar.Size, &bar.Pint,
		&bar.PriceChecked, &bar.Slug, &bar.Address, &bar.Fylke, &bar.FylkeName, &bar.FylkeSlug,
		&bar.Kommune, &bar.KommuneName, &bar.KommuneSlug, &bar.Sted, &bar.StedName,
		&bar.StedSlug, &bar.Flyplass, &bar.Brewery, &bar.Latitude, &bar.Longitude, &bar.TimedPrices, &bar.OrgNummer); err != nil {
		return bar, fmt.Errorf("db scan: %w", err)
	}

	return bar, nil
}

func UpdateHistoricPrice(conn *pgxpool.Pool, p models.Price) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
		INSERT INTO price_history
			(bar_id, price, size, pint, valid_from)
		VALUES
			($1, $2, $3, $4, $5);
	`

	cmdTag, err := conn.Exec(ctx, query,
		p.BarID,
		p.Price,
		p.Size,
		p.Pint,
		p.PriceUpdated)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}

func UpdateBarData(conn *pgxpool.Pool, p models.BarUpdateForm) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
        UPDATE bars
        SET
            bar = $1,
            brewery = $2,
            timed_prices = $3,
            address = $4,
            latitude = $5,
			longitude = $6,
			orgnummer = $7,
			slug = $8,
			is_active = $9
        WHERE id = $10;
    `

	cmdTag, err := conn.Exec(ctx, query,
		p.Name,
		p.Brewery,
		p.TimedPrices,
		p.Address,
		p.Latitude,
		p.Longitude,
		p.OrgNummer,
		p.Slug,
		p.IsActive,
		p.ID)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id %d", p.ID)
	}

	return nil
}

func GetPendingPrices(ctx context.Context, pool *pgxpool.Pool) ([]models.UpdatedPrice, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT target_id, target_table, price, size, pint, price_updated, price_checked FROM price_control ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	prices, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.UpdatedPrice])

	if err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return prices, nil
}
