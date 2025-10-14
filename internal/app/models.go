package app

import (
	"go-router/internal/bars"
	"go-router/internal/prices"
)

type CreateBarForm struct {
	Bar   bars.BarManual
	Price prices.Price
}
