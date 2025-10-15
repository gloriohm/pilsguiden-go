package prices

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateNewPrice(ctx context.Context, conn *pgxpool.Pool, input Price) error {
	priceAutoFormat(&input)
	err := createNewPriceRow(ctx, conn, input)
	return err
}

func CreateNewPriceControl(ctx context.Context, conn *pgxpool.Pool, input PriceControl) error {
	input.Pint = ToPint(input.Price, input.Size)
	input.Reported = time.Now()
	err := createNewPriceControlRow(ctx, conn, input)
	return err
}

func UpdatePrice(ctx context.Context, conn *pgxpool.Pool, input Price) error {
	priceAutoFormat(&input)
	if err := updatePriceHistory(ctx, conn, input.ID); err != nil {
		return err
	}
	err := updatePriceRow(ctx, conn, input)
	return err
}

func priceAutoFormat(p *Price) {
	if p.Size != 0.5 {
		p.Pint = ToPint(p.Price, p.Size)
	} else {
		p.Pint = p.Price
	}
	now := time.Now()

	p.PriceChecked = now
	p.PriceUpdated = now
}

func ToPint(price int, size float32) int {
	var pint int
	if size == 0.5 {
		pint = price
	} else {
		pint = int(float32(price) / size / 2)
	}
	return pint
}
