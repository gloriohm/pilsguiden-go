package app

import (
	"context"
	"go-router/internal/bars"
	"go-router/internal/prices"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo interface {
	List(ctx context.Context, limit int) ([]Bar, error)
	ByID(ctx context.Context, id int64) (Bar, error)
	Create(ctx context.Context, b Bar) (int64, error)
}

type Service struct{ repo Repo }

func NewService(r Repo) *Service { return &Service{repo: r} }

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
