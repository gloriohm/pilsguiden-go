package bars

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createBarRow(ctx context.Context, conn *pgxpool.Pool, bar Bar) (int, error) {
	query := `INSERT INTO bars (name, address, brewery, orgnummer, osm_id, linked_bar, slug, fylke, kommune, sted, latitude, longitude) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query,
		bar.Name,
		bar.Address,
		bar.Brewery,
		bar.OrgNummer,
		bar.OsmID,
		bar.LinkedBar,
		bar.Slug,
		bar.Fylke,
		bar.Kommune,
		bar.Sted,
		bar.Latitude,
		bar.Longitude).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("could not create bar: %w", err)
	}

	return id, nil
}

func createBarMetadata(ctx context.Context, conn *pgxpool.Pool, meta BarMetadata) {
	query := `INSERT INTO bars (bar_id, last_osm_sync, cuisine, opening_hours, wheelchair, website, email, phone, facebook, instagram) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	conn.QueryRow(ctx, query, meta.BarID, meta.LastOSMSync, meta.Cuisine, meta.OpeningHours, meta.Wheelchair, meta.Website, meta.Email, meta.Phone, meta.Facebook, meta.Instagram)
}

// LOCATION HELPERS
func GetLocationIdByName(ctx context.Context, conn *pgxpool.Pool, name, level string) (int, error) {
	query := `SELECT id FROM locs WHERE name = $1 AND hierarchy = $2`

	row := conn.QueryRow(ctx, query, name, level)

	var locID int
	if err := row.Scan(&locID); err != nil {
		return 0, fmt.Errorf("db scan: %w", err)
	}

	return locID, nil
}

func createNewLocation(ctx context.Context, conn *pgxpool.Pool, loc Location) (int, error) {
	query := `INSERT INTO locs (name, slug, hierarchy, parent) VALUES ($1, $2, $3, $4) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, loc.Name, loc.Slug, loc.Hierarchy, loc.Parent).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("could not create location: %w", err)
	}

	return id, nil
}
