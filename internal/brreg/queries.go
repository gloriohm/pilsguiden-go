package brreg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getOrgID(ctx context.Context, conn *pgxpool.Pool, table, orgnum string) (int, error) {
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

func createHovedenhet(ctx context.Context, conn *pgxpool.Pool, data Hovedenhet) error {
	query := `INSERT INTO organisasjoner (name, orgnummer, adresse, postnummer, sted, kommune, kommunenummer, konkurs, under_avvikling, under_tvangsavvikling, stiftelsesdato) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, data.Name, data.Orgnummer, data.Adresse, data.Postnummer, data.Sted, data.Kommune, data.Kommunenummer, data.Konkurs, data.UnderAvvikling, data.UnderTvangsavvikling, data.Stiftelsesdato).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func createUnderenhet(ctx context.Context, conn *pgxpool.Pool, data Underenhet) error {
	query := `INSERT INTO underenheter (name, orgnummer, parent, adresse, postnummer, sted, kommune, kommunenummer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	var id int
	err := conn.QueryRow(ctx, query, data.Name, data.Orgnummer, data.Parent, data.Adresse, data.Postnummer, data.Sted, data.Kommune, data.Kommunenummer).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}
