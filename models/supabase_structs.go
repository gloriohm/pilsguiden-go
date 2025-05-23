package models

import "time"

type Bar struct {
	BarManual
	BarAutoFormat
	BarOSM
}

type BarMetadata struct {
	BarID        int       `json:"bar_id"`
	LastOSMSync  time.Time `json:"last_osm_sync"`
	LinkedBar    bool      `json:"linked_bar"`
	Type         string    `json:"type"`
	Cuisine      *string   `json:"cuisine"`
	OpeningHours *string   `json:"opening_hours"`
	Wheelchair   *string   `json:"wheelchair"`
	Website      *string   `json:"website"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	Facebook     *string   `json:"facebook"`
	Instagram    *string   `json:"instagram"`
}

type BarManual struct {
	Name      string  `json:"name" form:"name"`
	Address   string  `json:"address" form:"address"`
	Flyplass  *int16  `json:"flyplass" form:"flyplass"`
	Price     int16   `json:"price" form:"price"`
	Size      float64 `json:"size" form:"size"`
	Brewery   *string `json:"brewery" form:"brewery"`
	OrgNummer string  `json:"orgnummer" form:"orgnummer"`
	OsmID     string  `json:"osm_id" form:"osm_id"`
	LinkedBar bool    `json:"linked_bar" form:"linked_bar"`
}

type BarAutoFormat struct {
	Pint         int16     `json:"pint"`
	Slug         string    `json:"slug"`
	PriceUpdated time.Time `json:"price_updated"`
	PriceChecked time.Time `json:"price_checked"`
	IsActive     bool      `json:"is_active"`
	TimedPrices  bool      `json:"timed_prices"`
}

type BarOSM struct {
	Fylke     int     `json:"fylke"`
	Kommune   int     `json:"kommune"`
	Sted      *int    `json:"sted"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
