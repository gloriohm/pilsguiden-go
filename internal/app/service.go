package app

import (
	"context"
	"go-router/internal/bars"
	"go-router/internal/prices"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateBarWithPrice(ctx context.Context, conn *pgxpool.Pool, input CreateBarForm) error {
	barID, err := bars.CreateBar(ctx, conn, input.Bar)
	if err != nil {
		return err
	}

	input.Price.BarID = barID

	err = prices.CreateNewPrice(ctx, conn, input.Price)
	if err != nil {
		return err
	}

	return nil
}
