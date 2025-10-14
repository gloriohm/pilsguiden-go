package osm

type nodeDetails struct {
	Lat         string    `json:"lat"`
	Lon         string    `json:"lon"`
	DisplayName string    `json:"display_name"`
	Address     address   `json:"address"`
	ExtraTags   extraTags `json:"extratags"`
	Type        string    `json:"type"`
}

type addressParts struct {
	Sted     string
	Kommune  string
	Fylke    string
	Postcode string
	Lat      float64
	Lon      float64
}

type address struct {
	CountryCode  string `json:"country_code"`
	Road         string `json:"road"`
	Postcode     string `json:"postcode"`
	Suburb       string `json:"suburb"`
	Municipality string `json:"municipality"`
}

type extraTags struct {
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
