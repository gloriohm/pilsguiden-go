package bars

import (
	"go-router/internal/osm"
	"time"
)

type BarManual struct {
	Name      string `db:"bar" form:"name"`
	Address   string `db:"address" form:"address"`
	Brewery   string `db:"brewery" form:"brewery"`
	OrgNummer string `db:"orgnummer" form:"orgnummer"`
	OsmID     string `db:"osm_id" form:"osm_id"`
	LinkedBar bool   `db:"linked_bar" form:"linked_bar"`
}

type BarLocation struct {
	Fylke     int     `db:"fylke"`
	Kommune   int     `db:"kommune"`
	Sted      *int    `db:"sted"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
}

type Bar struct {
	BarManual
	BarLocation
	Slug string
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

func ToBase(in []Location) []BaseLocation {
	out := make([]BaseLocation, len(in))
	for i := range in {
		out[i] = in[i].BaseLocation
	}
	return out
}

func ExtractBarMetadata(id int, nd *osm.NodeDetails) BarMetadata {
	return BarMetadata{
		BarID:        id,
		Type:         nd.Type,
		Cuisine:      nd.ExtraTags.Cuisine,
		OpeningHours: nd.ExtraTags.OpeningHours,
		Wheelchair:   nd.ExtraTags.Wheelchair,
		Website:      nd.ExtraTags.Website,
		Email:        nd.ExtraTags.Email,
		Phone:        nd.ExtraTags.Phone,
		Facebook:     nd.ExtraTags.Facebook,
		Instagram:    nd.ExtraTags.Instagram,
		LastOSMSync:  time.Now(),
	}
}
