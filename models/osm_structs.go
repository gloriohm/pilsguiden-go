package models

type Address struct {
	CountryCode  string `json:"country_code"`
	Road         string `json:"road"`
	Postcode     string `json:"postcode"`
	Suburb       string `json:"suburb"`
	Municipality string `json:"municipality"`
}

type ExtraTags struct {
	LiveMusic    *string `json:"live_music"`
	Cuisine      *string `json:"cuisine"`
	Wheelchair   *string `json:"wheelchair"`
	OpeningHours *string `json:"opening_hours"`
	Website      *string `json:"website"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	Facebook     *string `json:"contact:facebook"`
	Instagram    *string `json:"contact:instagram"`
}

type NodeDetails struct {
	Lat         string    `json:"lat"`
	Lon         string    `json:"lon"`
	DisplayName string    `json:"display_name"`
	Address     Address   `json:"address"`
	ExtraTags   ExtraTags `json:"extratags"`
	Type        string    `json:"type"`
}

type AddressParts struct {
	Sted     string
	Kommune  string
	Fylke    string
	Postcode string
	Lat      float64
	Lon      float64
}

type BarData struct {
	NodeDetails
	OSMNode       string
	FormattedAddr string
	Street        string
	LinkedBar     bool
	LiveMusic     bool
	Food          bool
}

type AddrIDs struct {
	Fylke   int
	Kommune int
	Sted    int
}
