package models

import "time"

type Preferences struct {
	CustomTime bool
	Time       time.Time
	Date       string
}

type Navigation struct {
	Level string
	ID    int
}

type SessionStore struct {
	Navigation  Navigation
	Preferences Preferences
	BarsFilter  BarsFilter
}

type RawCustomTime struct {
	Time string `form:"customTime"`
	Day  int    `form:"customDay"`
}

type UpdateBarStore struct {
	BarID int
	Price int
	Size  float64
}

type BarsFilter struct {
	Order     int
	Breweries []string
	MaxPrice  *int
	MinPrice  *int
}
