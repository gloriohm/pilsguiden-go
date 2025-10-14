package prices

import (
	"time"

	"github.com/jackc/pgtype"
)

type Price struct {
	ID           int       `db:"id"`
	BarID        int       `db:"bar_id"`
	Price        int       `db:"price" form:"price"`
	Pint         int       `db:"pint"`
	Size         float32   `db:"size" form:"size"`
	PriceUpdated time.Time `db:"price_updated"`
	PriceChecked time.Time `db:"price_checked"`
	DefaultPrice bool      `db:"default_price"`
}

type PriceTime struct {
	ID             int         `db:"id"`
	PriceID        int         `db:"price_id"`
	FromTime       time.Time   `db:"from_time"`
	UntilTime      time.Time   `db:"until_time"`
	Day            int         `db:"day"`
	PassesMidnight bool        `db:"passes_midnight"`
	EndDay         pgtype.Int8 `db:"end_day"` // null when passes_midnight is false
}

type PriceControl struct {
	PriceID  int
	Price    int
	Size     float32
	Pint     int
	Reported time.Time
}
