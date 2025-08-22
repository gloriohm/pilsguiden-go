package models

import (
	"time"

	"github.com/jackc/pgtype"
)

type Bar struct {
	ID int `db:"id"`
	BarManual
	BarAutoFormat
	BarOSM
	BarRelativeData
	BarExpandedLocation
}

type BarView struct {
	ID           int         `db:"id"`
	Name         string      `db:"bar" form:"name"`
	Address      string      `db:"address"`
	Price        int         `db:"price"`
	Size         float64     `db:"size"`
	Pint         int         `db:"pint"`
	Brewery      string      `db:"brewery"`
	Slug         string      `db:"slug"`
	PriceChecked time.Time   `db:"price_checked"`
	Flyplass     pgtype.Int8 `db:"flyplass"`
	BarOSM
	BarExpandedLocation
	BarRelativeData
}

type BarRelativeData struct {
	CurrentPint  int         `db:"current_pint"`
	CurrentPrice int         `db:"current_price"`
	FromTime     *time.Time  `db:"from_time"`
	UntilTime    *time.Time  `db:"until_time"`
	HappyChecked *time.Time  `db:"hk_checked"`
	HkeyID       pgtype.Int8 `db:"hkey_id"`
}

type BarLocationNames struct {
	FylkeName   string
	KommuneName string
	StedName    *string
}

type BarExpandedLocation struct {
	BarLocationNames
	FylkeSlug   string
	KommuneSlug string
	StedSlug    *string
}

type BarMetadata struct {
	BarID        int       `db:"bar_id"`
	LastOSMSync  time.Time `db:"last_osm_sync"`
	LinkedBar    bool      `db:"linked_bar"`
	Type         string    `db:"type"`
	Cuisine      *string   `db:"cuisine"`
	OpeningHours *string   `db:"opening_hours"`
	Wheelchair   *string   `db:"wheelchair"`
	Website      *string   `db:"website"`
	Email        *string   `db:"email"`
	Phone        *string   `db:"phone"`
	Facebook     *string   `db:"facebook"`
	Instagram    *string   `db:"instagram"`
}

type BarManual struct {
	Name      string      `db:"bar" form:"name"`
	Address   string      `db:"address" form:"address"`
	Flyplass  pgtype.Int8 `db:"flyplass" form:"flyplass"`
	Price     int         `db:"price" form:"price"`
	Size      float64     `db:"size" form:"size"`
	Brewery   string      `db:"brewery" form:"brewery"`
	OrgNummer string      `db:"orgnummer" form:"orgnummer"`
	OsmID     string      `db:"osm_id" form:"osm_id"`
	LinkedBar bool        `db:"linked_bar" form:"linked_bar"`
}

type BarAutoFormat struct {
	Pint         int       `db:"pint"`
	Slug         string    `db:"slug"`
	PriceUpdated time.Time `db:"price_updated"`
	PriceChecked time.Time `db:"price_checked"`
	IsActive     bool      `db:"is_active"`
	TimedPrices  bool      `db:"timed_prices"`
}

type BarOSM struct {
	Fylke     int     `db:"fylke"`
	Kommune   int     `db:"kommune"`
	Sted      *int    `db:"sted"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
}

type AboutInfo struct {
	Total    int
	MaxPrice int
	MinPrice int
	Diff     int
}

type BaseLocation struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

type Location struct {
	BaseLocation
	Hierarchy string `db:"hierarchy"`
	Parent    *int   `db:"parent"`
}

type Brewery struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	Popular bool   `db:"popular"`
}

type UrlNav struct {
	Fylke   BaseLocation
	Kommune BaseLocation
	Sted    BaseLocation
}

type CurrentLvl struct {
	Name string
	ID   int
}

type HappyKey struct {
	ID             int         `db:"id"`
	BarID          int         `db:"bar"`
	Price          int         `db:"price"`
	Size           float64     `db:"size"`
	Pint           int         `db:"pint"`
	FromTime       time.Time   `db:"from_time"`
	UntilTime      time.Time   `db:"until_time"`
	Day            int         `db:"day"`
	PriceUpdated   time.Time   `db:"updated_at"`
	PriceChecked   time.Time   `db:"price_checked"`
	PassesMidnight bool        `db:"passes_midnight"`
	EndDay         pgtype.Int8 `db:"end_day"`
}

type SearchResult struct {
	ID   int
	Name string
	Slug string
	Type string
}

type User struct {
	ID    int
	Name  string
	Email string
	Admin bool
}

type UpdatedPrice struct {
	TargetID     int       `db:"target_id"`
	TargetTable  string    `db:"target_table"`
	Price        int       `db:"price"`
	Size         float64   `db:"size"`
	Pint         int       `db:"pint"`
	PriceUpdated time.Time `db:"price_updated"`
	PriceChecked time.Time `db:"price_checked"`
}

type BarUpdateForm struct {
	ID          int     `form:"id"`
	Name        string  `db:"bar" form:"name"`
	Price       int     `db:"price" form:"price"`
	Size        float64 `db:"size" form:"size"`
	Brewery     string  `db:"brewery" form:"brewery"`
	TimedPrices bool    `db:"timed_prices" form:"timed"`
	Address     string  `db:"address" form:"address"`
	Latitude    float64 `db:"latitude" form:"lat"`
	Longitude   float64 `db:"longitude" form:"lon"`
	OrgNummer   string  `db:"orgnummer" form:"orgnummer"`
	Slug        string  `db:"slug" form:"slug"`
	IsActive    bool    `db:"is_active" form:"active"`
}

type Price struct {
	BarID        int       `db:"id"`
	Price        int       `db:"price" form:"price"`
	Size         float64   `db:"size" form:"size"`
	Pint         int       `db:"pint"`
	PriceUpdated time.Time `db:"price_updated"`
	PriceChecked time.Time `db:"price_checked"`
}
