package database

import (
	"context"
	"errors"
	"fmt"
	"go-router/internal/utils"
	"go-router/models"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetBarsByLocation(conn *pgx.Conn, id int, column string) ([]models.BarView, error) {
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

func GetBarsByLocationAndTime(conn *pgx.Conn, id int, column, date, customTime string) ([]models.BarView, error) {
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
func GetBarBySlug(conn *pgx.Conn, slug string) (*models.Bar, error) {
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

func GetTopTen(conn *pgx.Conn) ([]models.BarView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func GetBottomTen(conn *pgx.Conn) ([]models.BarView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func GetLocationIdByName(conn *pgx.Conn, name, level string) (int, error) {
	validLvl := utils.CheckValidLocationLevel(level)
	if !validLvl {
		return 0, errors.New("not a valid location level")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id FROM locs WHERE name = $1 AND hierarchy = $2`

	row := conn.QueryRow(ctx, query, name, level)

	var loc int
	if err := row.Scan(&loc); err != nil {
		return 0, fmt.Errorf("db scan: %w", err)
	}

	return loc, nil
}

func CreateNewLocation(conn *pgx.Conn, loc models.Location) (int, error) {
	validLvl := utils.CheckValidLocationLevel(loc.Hierarchy)
	if !validLvl {
		return 0, errors.New("not a valid location level")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO locs (name, slug, hierarchy, parent) VALUES ($1, $2, $3, $4) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, loc.Name, loc.Slug, loc.Hierarchy, loc.Parent).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("could not create location: %w", err)
	}

	return id, nil
}

func CreateNewBar(conn *pgx.Conn, bar models.Bar) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO bars (name, address, flyplass, price, size, brewery, orgnummer, osm_id, linked_bar, pint, slug, price_updated, price_checked, is_active, timed_prices, fylke, kommune, sted, latitude, longitude) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, bar.Name, bar.Address, bar.Flyplass, bar.Price, bar.Size, bar.Brewery, bar.OrgNummer, bar.OsmID, bar.LinkedBar, bar.Pint, bar.Slug, bar.PriceUpdated, bar.PriceChecked, bar.IsActive, bar.TimedPrices, bar.Fylke, bar.Kommune, bar.Sted, bar.Latitude, bar.Longitude).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("could not create bar: %w", err)
	}

	return id, nil
}

func CreateBarMetadata(conn *pgx.Conn, meta models.BarMetadata) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO bars (bar_id, last_osm_sync, cuisine, opening_hours, wheelchair, website, email, phone, facebook, instagram) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	conn.QueryRow(ctx, query, meta.BarID, meta.LastOSMSync, meta.Cuisine, meta.OpeningHours, meta.Wheelchair, meta.Website, meta.Email, meta.Phone, meta.Facebook, meta.Instagram)
}

func GetOrgID(conn *pgx.Conn, table, orgnum string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	valid := table == "organisasjoner" || table == "underenheter"
	if !valid {
		return 0, errors.New("not a valid table name")
	}

	query := `SELECT id FROM $1 WHERE orgnummer = $2`

	var id int
	err := conn.QueryRow(ctx, query, table, orgnum).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("orgnummer not found: %w", err)
	}
	return id, nil
}

func CreateHovedenhet(conn *pgx.Conn, data models.Hovedenhet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO organisasjoner (name, orgnummer, adresse, postnummer, sted, kommune, kommunenummer, konkurs, under_avvikling, under_tvangsavvikling, stiftelsesdato) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, data.Name, data.Orgnummer, data.Adresse, data.Postnummer, data.Sted, data.Kommune, data.Kommunenummer, data.Konkurs, data.UnderAvvikling, data.UnderTvangsavvikling, data.Stiftelsesdato).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func CreateUnderenhet(conn *pgx.Conn, data models.Underenhet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO underenheter (name, orgnummer, parent, adresse, postnummer, sted, kommune, kommunenummer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, data.Name, data.Orgnummer, data.Parent, data.Adresse, data.Postnummer, data.Sted, data.Kommune, data.Kommunenummer).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func GetBarMetadata(conn *pgx.Conn, barID int) (models.BarMetadata, error) {
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

func GetAllLocations(conn *pgx.Conn) ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var locs []models.Location

	query := `SELECT id, name, slug, hierarchy, parent FROM locs ORDER BY name`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return locs, err
	}

	for rows.Next() {
		var loc models.Location
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

func GetHappyKeysByBarID(conn *pgx.Conn, barID int) ([]models.HappyKey, error) {
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

func CreateBrewery(conn *pgx.Conn, newBrew string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO breweries (name, popular) VALUES ($1, false) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, newBrew).Scan(&id)

	return err
}

func UpdatePriceChecked(conn *pgx.Conn, table string, newTime time.Time, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allowedTables := map[string]bool{
		"bars":       true,
		"happy_keys": true,
	}
	if !allowedTables[table] {
		return fmt.Errorf("invalid table name")
	}

	query := fmt.Sprintf("UPDATE %s SET price_checked = $1 WHERE id = $2", table)
	cmdTag, err := conn.Exec(ctx, query, newTime, id)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id %d", id)
	}

	return nil
}

func GetSearchResult(conn *pgx.Conn, keyword string) ([]models.SearchResult, error) {
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

func UpdateBreweryWhereUnknown(conn *pgx.Conn, bar int, brew string) error {
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

func GetBarSearchResult(conn *pgx.Conn, keyword string) ([]models.SearchResult, error) {
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
