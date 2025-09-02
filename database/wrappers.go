package database

import (
	"go-router/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateCurrentAndHistoricPrice(conn *pgxpool.Pool, newPrice models.Price) error {
	oldPrice, err := GetPrice(conn, newPrice.BarID)
	if err != nil {
		return err
	}

	if err := UpdateHistoricPrice(conn, oldPrice); err != nil {
		return err
	}

	if err := UpdatePrice(conn, newPrice); err != nil {
		return err
	}

	return nil
}
