package prices

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createNewPriceRow(ctx context.Context, conn *pgxpool.Pool, data Price) error {
	const query = `
		INSERT INTO prices
			(bar_id, price, size, pint, price_updated, price_checked, default_price)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
	`
	cmdTag, err := conn.Exec(ctx, query,
		data.BarID,
		data.Price,
		data.Size,
		data.Pint,
		data.PriceUpdated,
		data.PriceChecked,
		data.DefaultPrice)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id")
	}

	return nil
}

func createNewPriceControlRow(ctx context.Context, conn *pgxpool.Pool, p PriceControl) error {
	const query = `
		INSERT INTO price_control
			(target_id, price, size, pint, price_reported)
		VALUES
			($1, $2, $3, $4, $5);
	`

	cmdTag, err := conn.Exec(ctx, query,
		p.PriceID,
		p.Price,
		p.Size,
		p.Pint,
		p.Reported)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id")
	}

	return nil
}

func UpdatePriceChecked(ctx context.Context, conn *pgxpool.Pool, newTime time.Time, id int) error {
	query := `UPDATE prices SET price_checked = $1 WHERE id = $2`
	cmdTag, err := conn.Exec(ctx, query, newTime, id)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id %d", id)
	}

	return nil
}

func UpdatePrice(ctx context.Context, conn *pgxpool.Pool, p Price) error {
	query := `
        UPDATE prices
        SET
            price          = $2,
            size           = $3,
            pint           = $4,
            price_updated  = $5,
            price_checked  = $6
        WHERE id = $1;
    `

	cmdTag, err := conn.Exec(ctx, query,
		p.BarID,
		p.Price,
		p.Size,
		p.Pint,
		p.PriceUpdated,
		p.PriceChecked)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for id %d", p.BarID)
	}

	return nil
}
